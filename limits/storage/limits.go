package storage

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/underbek/examples-go/limits/domain"
)

func (s *Storage) MatchLimits(ctx context.Context, currency string, meta domain.Attributes) ([]domain.Limit, error) {
	query, args, err := sq.Select(
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
	).From("limits").
		Where(sq.Eq{"deleted_at": nil}).
		Where(sq.Eq{"currency": currency}).
		Where("meta <@ ?", meta).
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

	limits, err := pgx.CollectRows[domain.Limit](rows, pgx.RowToStructByName[domain.Limit])
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("collect rows failed")

		return nil, err
	}

	return limits, nil
}

func (s *Storage) CleanupLimits(ctx context.Context) error {
	query, _, err := sq.Delete("limits").
		Where("deleted_at IS NOT NULL").
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
