package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	ctxerrors "github.com/underbek/examples-go/errors"
	"github.com/underbek/examples-go/limits/domain"
	goKitPgx "github.com/underbek/examples-go/storage/pgx"
)

func (s *service) createOperationEntities(ctx context.Context, info domain.OperationInfo) (uint64, []uint64, uint64, error) {
	now := s.timeProvider.Now()

	tx, err := s.db.Begin(ctx, &pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("begin transaction failed")

		return 0, nil, 0, err
	}

	defer func() {
		if err == nil {
			err = tx.Commit(ctx)
			if err != nil {
				s.logger.WithCtx(ctx).
					WithError(err).
					Error("commit transaction failed")
			}
		} else {
			err = tx.Rollback(ctx)
			if err != nil {
				s.logger.WithCtx(ctx).
					WithError(err).
					Error("rollback transaction failed")
			}
		}
	}()

	st := s.createStorage(tx)

	ctxID, err := st.CreateContext(ctx, info.Meta)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("create context failed")

		return 0, nil, 0, err
	}

	limits, err := st.MatchLimits(ctx, info.Amount.Currency, info.Meta)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("match limits failed")

		return 0, nil, 0, err
	}

	static, dynamic := splitLimits(limits)
	err = checkStaticLimits(static, info.Amount.Value)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("check static limits failed")

		return 0, nil, 0, err
	}

	operation := domain.Operation{
		ContextID: ctxID,
		Value:     info.Amount.Value,
		Currency:  info.Amount.Currency,
		Status:    domain.OperationStatusNew,
	}

	operation, err = st.CreateOperation(ctx, operation)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("create operation failed")

		return 0, nil, 0, err
	}

	if len(dynamic) == 0 {
		return ctxID, nil, operation.ID, nil
	}

	counters, err := s.GenerateCounters(ctx, dynamic, now)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("generate counters failed")

		return 0, nil, 0, err
	}

	counterIDs, err := st.CreateCountersIfNotExists(ctx, counters)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("create counters failed")

		return 0, nil, 0, err
	}

	err = st.LinkCountersToOperation(ctx, counterIDs, operation.ID)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("link counters to operation failed")

		return 0, nil, 0, err
	}

	return ctxID, counterIDs, operation.ID, nil
}

func (s *service) updateOperationEntities(ctx context.Context, info domain.AppendOperationInfo) ([]uint64, uint64, domain.Context, error) {
	now := s.timeProvider.Now()

	tx, err := s.db.Begin(ctx, &pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("begin transaction failed")

		return nil, 0, domain.Context{}, err
	}

	defer func() {
		if err == nil {
			err = tx.Commit(ctx)
			if err != nil {
				s.logger.WithCtx(ctx).
					WithError(err).
					Error("commit transaction failed")
			}
		} else {
			err = tx.Rollback(ctx)
			if err != nil {
				s.logger.WithCtx(ctx).
					WithError(err).
					Error("rollback transaction failed")
			}
		}
	}()

	st := s.createStorage(tx)

	oldContext, err := st.GetContextByID(ctx, info.ContextID)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("get context failed")

		return nil, 0, domain.Context{}, err
	}

	newContext := oldContext
	newContext.Meta, err = mergeEntities(newContext.Meta, info.Meta)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("merge context meta failed")

		return nil, 0, domain.Context{}, err
	}

	if entitiesIsEqual(oldContext.Meta, newContext.Meta) {
		return nil, 0, newContext, nil
	}

	operations, err := st.GetOperationsByContextID(ctx, info.ContextID)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("get operations failed")

		return nil, 0, domain.Context{}, err
	}

	oldOperation, err := validateExistedOperations(operations)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			With("context_id", info.ContextID).
			Error("validate existed operations failed")

		return nil, 0, domain.Context{}, err
	}

	oldLimits, err := st.MatchLimits(ctx, oldOperation.Currency, oldContext.Meta)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("match limits failed")

		return nil, 0, domain.Context{}, err
	}

	newLimits, err := st.MatchLimits(ctx, oldOperation.Currency, newContext.Meta)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("match limits failed")

		return nil, 0, domain.Context{}, err
	}

	limits := filterOldLimits(newLimits, oldLimits)
	static, dynamic := splitLimits(limits)
	err = checkStaticLimits(static, oldOperation.Value)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("check static limits failed")

		return nil, 0, domain.Context{}, err
	}

	operation := domain.Operation{
		ContextID: newContext.ID,
		Value:     oldOperation.Value,
		Currency:  oldOperation.Currency,
		Status:    domain.OperationStatusNew,
	}

	operation, err = st.CreateOperation(ctx, operation)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("create operation failed")

		return nil, 0, domain.Context{}, err
	}

	if len(dynamic) == 0 {
		return nil, operation.ID, newContext, nil
	}

	counters, err := s.GenerateCounters(ctx, dynamic, now)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("generate counters failed")

		return nil, 0, domain.Context{}, err
	}

	counterIDs, err := st.CreateCountersIfNotExists(ctx, counters)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("create counters failed")

		return nil, 0, domain.Context{}, err
	}

	err = st.LinkCountersToOperation(ctx, counterIDs, operation.ID)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("link counters to operation failed")

		return nil, 0, domain.Context{}, err
	}

	return counterIDs, operation.ID, newContext, nil
}

func (s *service) incrementCountersAndUpdateContext(
	ctx context.Context,
	operationID uint64,
	counterIDs []uint64,
	domainContext *domain.Context,
) error {
	return s.transaction(ctx, &pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx goKitPgx.Transaction) error {
		st := s.createStorage(tx)

		var exceeded []domain.ExceededCounters
		var err error
		if domainContext != nil {
			exceeded, err = st.IncrementCountersAndUpdateContext(ctx, operationID, counterIDs, *domainContext)
		} else {
			exceeded, err = st.IncrementCounters(ctx, operationID, counterIDs)
		}

		if err != nil {
			s.logger.WithCtx(ctx).
				WithError(err).
				Error("increment counters failed")

			return err
		}

		if len(exceeded) > 0 {
			counter := exceeded[0]

			s.logger.WithCtx(ctx).
				With("exceeded_counters", exceeded).
				With("counter_id", counter.CounterID).
				With("limit_id", counter.LimitID).
				Error("operation exceeded limits")

			msg := fmt.Sprintf(
				"new counter value %s is greater than %s value %s limit_id %d",
				counter.NewValue,
				counter.LimitType,
				counter.LimitValue,
				counter.LimitID,
			)

			return ctxerrors.New(ctxerrors.TypeInternal, msg)

		}

		return nil
	})
}
