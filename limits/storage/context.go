package storage

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	ctxerrors "github.com/underbek/examples-go/errors"
	"github.com/underbek/examples-go/limits/domain"
)

func (s *Storage) CreateContext(ctx context.Context, meta domain.Attributes) (uint64, error) {
	query, args, err := sq.Insert("context").
		Columns("meta").
		Values(meta).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("create query failed")

		return 0, err
	}

	var ctxID uint64
	err = s.ext.QueryRow(ctx, query, args...).Scan(&ctxID)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("query row failed")

		return 0, err
	}

	return ctxID, nil
}

func (s *Storage) GetContextByID(ctx context.Context, id uint64) (domain.Context, error) {
	query, args, err := sq.Select(
		"id",
		"meta",
	).
		From("context").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("create query failed")

		return domain.Context{}, err
	}

	rows, err := s.ext.Query(ctx, query, args...)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("query failed")

		return domain.Context{}, err
	}

	domainContext, err := pgx.CollectOneRow[domain.Context](rows, pgx.RowToStructByName[domain.Context])
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("collect rows failed")

		if err == pgx.ErrNoRows {
			return domainContext, ctxerrors.Wrap(
				err,
				ctxerrors.TypeNotFound,
				"no context found",
			)
		}

		return domain.Context{}, err
	}

	return domainContext, nil
}

func (s *Storage) UpdateContext(ctx context.Context, domainContext domain.Context) error {
	query, args, err := sq.Update("context").
		Set("meta", domainContext.Meta).
		Where(sq.Eq{"id": domainContext.ID}).
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

const deleteLimit = 10_000

func (s *Storage) CleanupContext(ctx context.Context, outdate time.Duration) error {
	timeOffset := time.Now().UTC().Add(-outdate)

	for offset := uint64(0); ; offset += deleteLimit {
		selectBuilder := sq.Select("c.id").
			From("context c").
			LeftJoin("operations o ON o.context_id = c.id").
			LeftJoin("operation_to_counter oc ON o.id = oc.operation_id").
			Where(sq.Expr("c.updated_at < ?", timeOffset)).
			Where(sq.Eq{"oc.counter_id": nil}).
			Limit(deleteLimit).
			Offset(offset).
			PlaceholderFormat(sq.Dollar)

		query, args, err := sq.Delete("context").
			Where(sq.Expr("id IN (?)", selectBuilder)).
			PlaceholderFormat(sq.Dollar).
			ToSql()
		if err != nil {
			return errors.Wrap(err, "sq.Delete")
		}

		command, err := s.ext.Exec(ctx, query, args...)
		if err != nil {
			return errors.Wrap(err, "ext.Exec")
		}

		s.logger.WithCtx(ctx).Debug(fmt.Sprintf("Removed %d rows on cleanup context", command.RowsAffected()))

		if command.RowsAffected() == 0 {
			return nil
		}
	}
}
