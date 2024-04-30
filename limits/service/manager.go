package service

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/underbek/examples-go/limits/domain"
)

func (s *service) CreateLimit(ctx context.Context, limit domain.Limit) (domain.Limit, error) {
	err := validateLimit(&limit)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("validate limit failed")

		return domain.Limit{}, err
	}

	domain.SortEntities(limit.Entities)

	limit.Hash = generateLimitHash(limit)

	limit, err = s.createStorage(s.db).CreateLimit(ctx, limit)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("create limit failed")

		return domain.Limit{}, err
	}

	return limit, nil
}

func (s *service) DeleteLimits(ctx context.Context, ids []uint64) error {
	var scopeIDs []uint64

	tx, err := s.db.Begin(ctx, &pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("begin transaction failed")

		return err
	}

	defer func() {
		err = tx.Rollback(ctx)
		if err != nil {
			s.logger.WithCtx(ctx).
				WithError(err).
				Error("rollback transaction failed")
		}
	}()

	st := s.createStorage(tx)
	err = s.deleteLimits(ctx, st, ids, scopeIDs)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("commit transaction failed")

		return err
	}

	return nil
}

func (s *service) UpdateLimit(ctx context.Context, limit domain.Limit) (domain.Limit, error) {
	err := validateLimit(&limit)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("validate limit failed")

		return domain.Limit{}, err
	}

	domain.SortEntities(limit.Entities)

	limit.Hash = generateLimitHash(limit)

	tx, err := s.db.Begin(ctx, &pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("begin transaction failed")

		return domain.Limit{}, err
	}

	defer func() {
		err = tx.Rollback(ctx)
		if err != nil {
			s.logger.WithCtx(ctx).
				WithError(err).
				Error("rollback transaction failed")
		}
	}()

	st := s.createStorage(tx)
	storedLimit, err := st.GetLimitByID(ctx, limit.ID)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			With("limit_id", limit.ID).
			Error("get limit by id failed")

		return domain.Limit{}, err
	}

	if isLimitCanBeUpdated(limit, storedLimit) {
		s.logger.WithCtx(ctx).Debug("updating old limit")

		storedLimit.Value = limit.Value
		limit, err = st.UpdateLimitValue(ctx, storedLimit)
		if err != nil {
			s.logger.WithCtx(ctx).
				WithError(err).
				Error("update limit value failed")

			return domain.Limit{}, err
		}
	} else {
		s.logger.WithCtx(ctx).Debug("delete old limit and create new to update")

		err = s.deleteLimits(ctx, st, []uint64{limit.ID}, nil)
		if err != nil {
			return domain.Limit{}, err
		}

		limit, err = st.CreateLimit(ctx, limit)
		if err != nil {
			s.logger.WithCtx(ctx).
				WithError(err).
				Error("create limit failed")

			return domain.Limit{}, err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("commit transaction failed")

		return domain.Limit{}, err
	}

	return limit, nil
}

func isLimitCanBeUpdated(newLimit, oldLimit domain.Limit) bool {
	sameHash := newLimit.Hash == oldLimit.Hash
	sameTZ := newLimit.Timezone == nil && oldLimit.Timezone == nil

	if newLimit.Timezone != nil && oldLimit.Timezone != nil {
		sameTZ = *newLimit.Timezone == *oldLimit.Timezone
	}

	return sameHash && sameTZ
}

func (s *service) GetLimits(ctx context.Context, filter domain.LimitsFilter) ([]domain.Limit, uint64, error) {
	return s.createStorage(s.db).GetLimits(ctx, filter)
}

func (s *service) GetLimitByID(ctx context.Context, id uint64) (domain.Limit, error) {
	return s.createStorage(s.db).GetLimitByID(ctx, id)
}
