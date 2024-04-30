package storage

import (
	"context"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/underbek/examples-go/limits/domain"
)

func (s *Storage) CreateCountersIfNotExists(ctx context.Context, counters []domain.Counter) ([]uint64, error) {
	builder := sq.Insert("counters").
		Columns(
			"hash",
			"limit_id",
			"value",
			"start_time",
			"end_time",
		)

	for _, counter := range counters {
		builder = builder.Values(
			counter.Hash,
			counter.LimitID,
			"0",
			counter.StartTime,
			counter.EndTime,
		)
	}

	query, args, err := builder.
		Suffix("ON CONFLICT DO NOTHING").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("create insert query failed")

		return nil, err
	}

	_, err = s.ext.Exec(ctx, query, args...)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("insert exec failed")

		return nil, err
	}

	hashes := make([]string, 0, len(counters))
	for _, counter := range counters {
		hashes = append(hashes, counter.Hash)
	}

	query, args, err = sq.Select("id").
		From("counters").
		Where(sq.Eq{"hash": hashes}).
		Where(sq.Eq{"deleted_at": nil}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("create select query failed")

		return nil, err
	}

	rows, err := s.ext.Query(ctx, query, args...)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("select query failed")

		return nil, err
	}

	ids, err := pgx.CollectRows[uint64](rows, pgx.RowTo[uint64])
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("collect rows failed")
	}

	return ids, nil
}

func (s *Storage) IncrementCounters(ctx context.Context, operationID uint64, counterIDs []uint64) ([]domain.ExceededCounters, error) {
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
                WHEN ((l.limit_type = 'total_count')) THEN (c.value::bigint + 1)::varchar
                WHEN ((l.limit_type = 'total_amount')) THEN (c.value::numeric + o.value::numeric)::varchar
               END)     AS new_value
    FROM counters c
             JOIN limits l ON c.limit_id = l.id
             JOIN operations o ON o.id = $1
    WHERE c.id = ANY($2)
),
     updated_counters AS (
         UPDATE counters c
             SET value = CASE
                             WHEN (
                                 (counter_info.new_value::numeric <= counter_info.limit_value::numeric)
                                 ) THEN counter_info.new_value
                             ELSE value
                 END,
				 updated_at = now()
             FROM counter_info
             WHERE c.id = counter_info.counter_id
             RETURNING counter_id, counter_info.limit_id, limit_type, period, meta, new_value, limit_value
     ),
     updated_operations AS (
         UPDATE operations o
             SET status = 'pending',
				 updated_at = now()
             WHERE o.id = $1
     )
SELECT counter_id, limit_id, limit_type, period, meta, new_value, limit_value
FROM updated_counters
WHERE limit_value::numeric < new_value::numeric;`

	rows, err := s.ext.Query(ctx, query, operationID, counterIDs)
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

func (s *Storage) IncrementCountersAndUpdateContext(
	ctx context.Context,
	operationID uint64,
	counterIDs []uint64,
	domainContext domain.Context,
) ([]domain.ExceededCounters, error) {
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
                WHEN ((l.limit_type = 'total_count')) THEN (c.value::bigint + 1)::varchar
                WHEN ((l.limit_type = 'total_amount')) THEN (c.value::numeric + o.value::numeric)::varchar
               END)     AS new_value
    FROM counters c
             JOIN limits l ON c.limit_id = l.id
             JOIN operations o ON o.id = $1
    WHERE c.id = ANY($2)
),
     updated_counters AS (
         UPDATE counters c
             SET value = CASE
                             WHEN (
                                 (counter_info.new_value::numeric <= counter_info.limit_value::numeric)
                                 ) THEN counter_info.new_value
                             ELSE value
                 END,
				 updated_at = now()
             FROM counter_info
             WHERE c.id = counter_info.counter_id
             RETURNING counter_id, counter_info.limit_id, limit_type, period, meta, new_value, limit_value
     ),
     updated_operations AS (
         UPDATE operations o
             SET status = 'pending',
				 updated_at = now()
             WHERE o.id = $1
     ),
	 updated_context AS (
		UPDATE context ctx
			SET meta = $3,
				updated_at = now()
			WHERE ctx.id = $4
	 )
SELECT counter_id, limit_id, limit_type, period, meta, new_value, limit_value
FROM updated_counters
WHERE limit_value::numeric < new_value::numeric;`

	rows, err := s.ext.Query(ctx, query, operationID, counterIDs, domainContext.Meta, domainContext.ID)
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

func (s *Storage) CleanupCounters(ctx context.Context, outdate time.Duration) error {
	hoursDuration := strconv.FormatInt(int64(outdate.Hours()), 10)

	query, _, err := sq.Delete("counters").
		Where("deleted_at IS NOT NULL OR end_time < (CURRENT_TIMESTAMP - INTERVAL '" + hoursDuration + " HOURS')").
		ToSql()
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("failed to create query")
		return err
	}

	if _, err = s.ext.Exec(ctx, query); err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("failed to execute query")
		return err
	}

	return nil
}
