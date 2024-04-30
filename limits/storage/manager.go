package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	ctxerrors "github.com/underbek/examples-go/errors"
	"github.com/underbek/examples-go/limits/domain"
)

func (s *Storage) CreateLimit(ctx context.Context, limit domain.Limit) (domain.Limit, error) {
	query, args, err := sq.Insert("limits").
		Columns(
			"hash",
			"currency",
			"meta",
			"limit_type",
			"value",
			"period",
			"timezone",
		).
		Values(
			limit.Hash,
			limit.Currency,
			limit.Entities,
			limit.LimitType,
			limit.Value,
			limit.Period,
			limit.Timezone,
		).
		Suffix("RETURNING id, created_at, updated_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("create query failed")

		return domain.Limit{}, err
	}

	rows, err := s.ext.Query(ctx, query, args...)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("query failed")

		return domain.Limit{}, err
	}

	type createdData struct {
		ID        uint64    `db:"id"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}

	result, err := pgx.CollectOneRow[createdData](rows, pgx.RowToStructByName[createdData])
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("collect one row failed")

		if strings.Contains(err.Error(), "duplicate key value violates unique") {
			return domain.Limit{}, ctxerrors.Wrap(
				err,
				ctxerrors.TypeInvalidRequest,
				fmt.Sprintf("limit with hash: \"%s\" already exist", limit.Hash),
			)
		}

		return domain.Limit{}, err
	}

	limit.ID = result.ID
	limit.CreatedAt = result.CreatedAt
	limit.UpdatedAt = result.UpdatedAt

	return limit, nil
}

func (s *Storage) DeleteLimits(ctx context.Context, ids []uint64) ([]uint64, error) {
	builder := sq.Update("limits").
		Set("deleted_at", "now()").
		Where(sq.Eq{"id": ids}).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar)

	query, args, err := builder.ToSql()
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
			Error("exec failed")

		return nil, err
	}

	settings, err := pgx.CollectRows(rows, pgx.RowTo[uint64])
	if err != nil {
		return nil, err
	}

	return settings, nil
}

func (s *Storage) DeleteCounters(ctx context.Context, limitIDs []uint64) error {
	query, args, err := sq.Update("counters").
		Set("deleted_at", "now()").
		Where(sq.Eq{"limit_id": limitIDs}).
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

func (s *Storage) GetLimitByID(ctx context.Context, id uint64) (domain.Limit, error) {
	builder := sq.Select(
		"id",
		"hash",
		"currency",
		"meta",
		"limit_type",
		"value",
		"period",
		"timezone",
		"created_at",
		"updated_at",
	).
		From("limits").
		Where(sq.Eq{"id": id}).
		Where(sq.Eq{"deleted_at": nil}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := builder.ToSql()
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("create query failed")

		return domain.Limit{}, err
	}

	rows, err := s.ext.Query(ctx, query, args...)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("query failed")

		return domain.Limit{}, err
	}

	limit, err := pgx.CollectOneRow[domain.Limit](rows, pgx.RowToStructByName[domain.Limit])
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("collect one row failed")

		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Limit{}, ctxerrors.Wrap(
				err,
				ctxerrors.TypeInvalidRequest,
				"no limit by id found",
			)
		}

		return domain.Limit{}, err
	}

	return limit, nil
}

func (s *Storage) UpdateLimitValue(ctx context.Context, limit domain.Limit) (domain.Limit, error) {
	query, args, err := sq.Update("limits").
		Set("value", limit.Value).
		Set("updated_at", "now()").
		Where(sq.Eq{"id": limit.ID}).
		Where(sq.Eq{"deleted_at": nil}).
		Suffix("RETURNING updated_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("create query failed")

		return domain.Limit{}, err
	}

	rows, err := s.ext.Query(ctx, query, args...)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("query failed")

		return domain.Limit{}, err
	}

	limit.UpdatedAt, err = pgx.CollectOneRow[time.Time](rows, pgx.RowTo[time.Time])
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("collect one row failed")

		return domain.Limit{}, err
	}

	return limit, nil
}
