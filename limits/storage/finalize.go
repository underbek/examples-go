package storage

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/underbek/examples-go/limits/domain"
)

func (s *Storage) CommitOperations(ctx context.Context, ids []uint64) error {
	query, args, err := sq.Update("operations").
		Set("status", domain.OperationStatusCommitted).
		Set("updated_at", "now()").
		Where(sq.Eq{"id": ids}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		s.logger.WithCtx(ctx).
			With("ids", ids).
			WithError(err).
			Error("build query failed")

		return err
	}

	_, err = s.ext.Exec(ctx, query, args...)
	if err != nil {
		s.logger.WithCtx(ctx).
			With("ids", ids).
			WithError(err).
			Error("commit operations failed")

		return err
	}

	return nil
}

func (s *Storage) RollbackOperations(ctx context.Context, ids []uint64) ([]domain.ExceededCounters, error) {
	query := `WITH counter_info AS (
    SELECT c.id         AS counter_id,
           l.id         AS limit_id,
           l.limit_type AS limit_type,
           l.period     AS period,
           c.value      AS counter_value,
           l.value      AS limit_value,
           o.id         AS operation_id,
           l.meta       AS meta,
           (CASE
                WHEN ((l.limit_type = 'total_count')) THEN (c.value::bigint - 1)::varchar
                WHEN ((l.limit_type = 'total_amount')) THEN (c.value::numeric - o.value::numeric)::varchar
               END)     AS new_value
	FROM operations o
			JOIN operation_to_counter otc ON otc.operation_id = o.id
			JOIN counters c ON c.id = otc.counter_id
			JOIN limits l ON c.limit_id = l.id
    WHERE o.id = ANY($1)
	AND c.deleted_at IS NULL
),
     updated_counters AS (
         UPDATE counters c
             SET value = counter_info.new_value::numeric,
				 updated_at = now()
             FROM counter_info
             WHERE c.id = counter_info.counter_id
             RETURNING counter_id, counter_info.limit_id, limit_type, period, meta, new_value, limit_value
     ),
     updated_operations AS (
         UPDATE operations o
             SET status = $2,
				 updated_at = now()
			 WHERE o.id = ANY($1)
     )
SELECT counter_id, limit_id, limit_type, period, meta, new_value, limit_value
FROM updated_counters
WHERE new_value::numeric < 0;`

	rows, err := s.ext.Query(ctx, query, ids, domain.OperationStatusRollback)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("select query failed")

		return nil, err
	}

	result, err := pgx.CollectRows[domain.ExceededCounters](rows, pgx.RowToStructByName[domain.ExceededCounters])
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("collect rows failed")

		return nil, err
	}

	return result, err
}
