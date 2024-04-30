package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	ctxerrors "github.com/underbek/examples-go/errors"
	"github.com/underbek/examples-go/limits/domain"
	goKitPgx "github.com/underbek/examples-go/storage/pgx"
)

func (s *service) SendOperation(ctx context.Context, info domain.OperationInfo) (uint64, error) {
	err := validateOperationInfo(info)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("validate operation info failed")

		return 0, err
	}

	ctxID, counterIDs, operationID, err := s.createOperationEntities(ctx, info)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("create operation entities failed")

		return 0, err
	}

	err = s.incrementCountersAndUpdateContext(ctx, operationID, counterIDs, nil)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("increment counters failed")

		return 0, err
	}

	return ctxID, nil
}

func (s *service) AppendOperation(ctx context.Context, info domain.AppendOperationInfo) (uint64, error) {
	err := validateEntities(info.Meta)
	if err != nil {
		s.logger.WithCtx(ctx).
			With("context_id", info.ContextID).
			WithError(err).
			Error("validate entities failed")

		return 0, err
	}

	counterIDs, operationID, domainContext, err := s.updateOperationEntities(ctx, info)
	if err != nil {
		s.logger.WithCtx(ctx).
			With("context_id", info.ContextID).
			WithError(err).
			Error("update operation entities failed")

		return 0, err
	}

	err = s.incrementCountersAndUpdateContext(ctx, operationID, counterIDs, &domainContext)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			Error("increment counters failed")

		return 0, err
	}

	return info.ContextID, nil
}

func (s *service) FinalizeOperations(ctx context.Context, info domain.FinalizeOperationsInfo) error {
	if info.Status != domain.OperationStatusCommitted && info.Status != domain.OperationStatusRollback {
		s.logger.WithCtx(ctx).
			With("status", info.Status).
			With("context_id", info.ContextID).
			Error("invalid operation status")

		return domain.ErrInvalidOperationStatus
	}

	operations, err := s.createStorage(s.db).GetOperationsByContextID(ctx, info.ContextID)
	if err != nil {
		s.logger.WithCtx(ctx).
			With("context_id", info.ContextID).
			WithError(err).
			Error("get operations by context id failed")

		return err
	}

	filteredOperations, err := filterOperations(operations)
	if err != nil {
		s.logger.WithCtx(ctx).
			With("context_id", info.ContextID).
			WithError(err).
			Error("filter operations failed")

		return err
	}

	operationIDs := make([]uint64, 0, len(filteredOperations))
	for _, operation := range filteredOperations {
		operationIDs = append(operationIDs, operation.ID)
	}

	return s.transaction(ctx, &pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx goKitPgx.Transaction) error {
		st := s.createStorage(tx)

		switch info.Status {
		case domain.OperationStatusCommitted:
			err = st.CommitOperations(ctx, operationIDs)
			if err != nil {
				s.logger.WithCtx(ctx).
					With("context_id", info.ContextID).
					With("operation_ids", operationIDs).
					WithError(err).
					Error("commit operations failed")

				return err
			}
		case domain.OperationStatusRollback:
			var exceeded []domain.ExceededCounters
			exceeded, err = st.RollbackOperations(ctx, operationIDs)
			if err != nil {
				s.logger.WithCtx(ctx).
					WithError(err).
					Error("rollback operations failed")

				return err
			}

			if len(exceeded) > 0 {
				s.logger.WithCtx(ctx).
					With("exceeded_counters", exceeded).
					Error("exceeded counters")

				return ctxerrors.New(
					ctxerrors.TypeInternal,
					fmt.Sprintf(
						"rollback counter %s with id %d failed",
						exceeded[0].LimitType,
						exceeded[0].CounterID,
					),
				)
			}
		}

		return nil
	})
}
