package storage

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/underbek/examples-go/limits/domain"
)

func (s *Storage) CreateOperation(ctx context.Context, operation domain.Operation) (domain.Operation, error) {
	query, args, err := sq.Insert("operations").
		Columns(
			"context_id",
			"value",
			"currency",
			"status",
		).
		Values(
			operation.ContextID,
			operation.Value,
			operation.Currency,
			operation.Status,
		).
		Suffix("RETURNING id, created_at, updated_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("create query failed")

		return domain.Operation{}, err
	}

	rows, err := s.ext.Query(ctx, query, args...)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("query failed")

		return domain.Operation{}, err
	}

	type resultData struct {
		ID        uint64    `db:"id"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}

	result, err := pgx.CollectOneRow[resultData](rows, pgx.RowToStructByName[resultData])
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("collect rows failed")

		return domain.Operation{}, err
	}

	operation.ID = result.ID
	operation.CreatedAt = result.CreatedAt
	operation.UpdatedAt = result.UpdatedAt

	return operation, nil
}

func (s *Storage) GetOperationsByContextID(ctx context.Context, ctxID uint64) ([]domain.Operation, error) {
	query, args, err := sq.Select(
		"o.id",
		"o.context_id",
		"o.currency",
		"o.value",
		"o.status",
		"o.created_at",
		"o.updated_at",
	).
		From("operations AS o").
		Join("context AS c ON o.context_id = c.id").
		Where(sq.Eq{"c.id": ctxID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("create query failed")

		return nil, err
	}

	rows, err := s.ext.Query(ctx, query, args...)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("query failed")

		return nil, err
	}

	operations, err := pgx.CollectRows[domain.Operation](rows, pgx.RowToStructByName[domain.Operation])
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("collect rows failed")

		return nil, err
	}

	return operations, nil
}

func (s *Storage) LinkCountersToOperation(ctx context.Context, counterIDs []uint64, operationID uint64) error {
	builder := sq.Insert("operation_to_counter").
		Columns(
			"counter_id",
			"operation_id",
		)

	for _, counterID := range counterIDs {
		builder = builder.Values(
			counterID,
			operationID,
		)
	}

	query, args, err := builder.
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("create query failed")

		return err
	}

	_, err = s.ext.Exec(ctx, query, args...)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("exec failed")

		return err
	}

	return nil
}
