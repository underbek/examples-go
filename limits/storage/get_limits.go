package storage

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/underbek/examples-go/limits/domain"
)

func (s *Storage) GetLimits(ctx context.Context, filter domain.LimitsFilter, scope []uint64) ([]domain.Limit, uint64, error) {
	builder, err := s.createLimitsFilterBuilder(filter)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("create filter builder failed")

		return nil, 0, err
	}

	count, err := s.getLimitsCount(ctx, builder)
	if err != nil {
		return nil, 0, err
	}

	if filter.Limit != nil && *filter.Limit < uint64(s.maxLimit) {
		builder = builder.Limit(*filter.Limit)
	} else {
		builder = builder.Limit(uint64(s.maxLimit))
	}

	if filter.Offset != nil {
		builder = builder.Offset(*filter.Offset)
	}

	query, args, err := builder.Columns(
		"id",
		"hash",
		"limit_type",
		"currency",
		"value",
		"meta",
		"period",
		"timezone",
		"created_at",
		"updated_at",
	).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("create query failed")

		return nil, 0, err
	}

	rows, err := s.ext.Query(ctx, query, args...)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("query failed")

		return nil, 0, err
	}

	limits, err := pgx.CollectRows[domain.Limit](rows, pgx.RowToStructByName[domain.Limit])
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("collect rows failed")

		return nil, 0, err
	}

	return limits, count, nil
}

func (s *Storage) getLimitsCount(ctx context.Context, builder sq.SelectBuilder) (uint64, error) {
	query, args, err := builder.Columns("COUNT(*)").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("create query failed")

		return 0, err
	}

	rows, err := s.ext.Query(ctx, query, args...)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("query failed")

		return 0, err
	}

	count, err := pgx.CollectOneRow[uint64](rows, pgx.RowTo[uint64])
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("collect one row failed")

		return 0, err
	}

	return count, nil
}

func (s *Storage) createLimitsFilterBuilder(filter domain.LimitsFilter) (sq.SelectBuilder, error) {
	builder := sq.Select().
		From("limits").
		Where(sq.Eq{"deleted_at": nil})

	if len(filter.LimitTypes) != 0 {
		builder = builder.Where(map[string]interface{}{"limit_type": filter.LimitTypes})
	}

	if filter.Currency != nil {
		builder = builder.Where(sq.Eq{"currency": filter.Currency})
	}

	if len(filter.Entities) != 0 {
		data, err := filter.Entities.Value()
		if err != nil {
			return sq.SelectBuilder{}, err
		}

		builder = builder.Where("meta @> ?", data)
	}

	if filter.Period != nil {
		builder = builder.Where(sq.Eq{"period": filter.Period})
	}

	if filter.Timezone != nil {
		builder = builder.Where(sq.Eq{"timezone": filter.Timezone})
	}

	return builder, nil
}
