package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	ctxerrors "github.com/underbek/examples-go/errors"
	"github.com/underbek/examples-go/limits/domain"
	"github.com/underbek/examples-go/limits/storage"
	goKitPgx "github.com/underbek/examples-go/storage/pgx"
	"golang.org/x/exp/maps"
)

func (s *service) transaction(ctx context.Context, opts *pgx.TxOptions, f func(goKitPgx.Transaction) error) error {
	var err error
	var retry bool

	for i := 0; i <= s.storageTrxCfg.RetryAmount; i++ {
		if retry, err = func() (bool, error) {
			var tx goKitPgx.Transaction
			tx, err = s.db.Begin(ctx, opts)
			if err != nil {
				s.logger.WithCtx(ctx).
					WithError(err).
					Error("begin transaction failed")

				return false, err
			}

			defer func() {
				if err = tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
					s.logger.WithCtx(ctx).
						WithError(err).
						Error("rollback transaction failed")
				}
			}()

			if err = f(tx); err != nil {
				return false, err
			}

			if err = tx.Commit(ctx); err != nil {
				dbErr := new(pgconn.PgError)
				if errors.As(err, &dbErr) && dbErr.Code == storage.ErrCodeSerializationFailure {
					s.logger.WithCtx(ctx).
						WithError(err).
						Warn("commit transaction failed due to transaction serialization failure")
					return true, err
				}

				s.logger.WithCtx(ctx).
					WithError(err).
					Error("commit transaction failed")

				return false, err
			}

			return false, nil
		}(); !retry {
			break
		}

		<-time.After(s.storageTrxCfg.RetryDelay)
	}

	return err
}

func generateLimitHash(limit domain.Limit) string {
	value := fmt.Sprintf("%s:%s", limit.LimitType, limit.Currency)

	if limit.Period != nil {
		value = fmt.Sprintf("%s:%s", value, limit.Period)
	}

	for _, entity := range limit.Entities {
		value = fmt.Sprintf("%s:%s:%s", value, entity.Name, entity.Value)
	}

	return base64.StdEncoding.EncodeToString([]byte(value))
}

func generateStartEndPeriods(limit domain.Limit, now time.Time) (time.Time, time.Time, error) {
	if limit.Period == nil {
		return time.Time{}, time.Time{}, fmt.Errorf("period is nil for limit %d", limit.ID)
	}

	if limit.Timezone == nil {
		return time.Time{}, time.Time{}, fmt.Errorf("timezone is nil for limit %d", limit.ID)
	}

	lc, err := time.LoadLocation(*limit.Timezone)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("load location failed for limit %d: %w", limit.ID, err)
	}

	now = now.In(lc)

	switch *limit.Period {
	case domain.PeriodTypeCALENDARDAY:
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, lc).UTC()
		end := start.Add(24 * time.Hour)
		return start, end, nil

	case domain.PeriodTypeCALENDARWEEK:
		weekday := now.Weekday()
		if weekday == time.Sunday {
			weekday = 7
		}
		weekday -= 1

		start := time.Date(now.Year(), now.Month(), now.Day()-int(weekday), 0, 0, 0, 0, lc).UTC()
		end := start.AddDate(0, 0, 7)
		return start, end, nil

	case domain.PeriodTypeCALENDARMONTH:
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, lc).UTC()
		end := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, lc).UTC()
		return start, end, nil
	}

	return time.Time{}, time.Time{}, fmt.Errorf("unknown period type %s for limit %d", *limit.Period, limit.ID)
}

func GenerateCounterHash(limit domain.Limit, start time.Time) (string, error) {
	if limit.Period == nil {
		return "", fmt.Errorf("period is nil for limit %d", limit.ID)
	}

	value := fmt.Sprintf("%s:%s:%s:%s", limit.LimitType, limit.Currency, limit.Period, start)

	for _, entity := range limit.Entities {
		value = fmt.Sprintf("%s:%s:%s", value, entity.Name, entity.Value)
	}

	return base64.StdEncoding.EncodeToString([]byte(value)), nil
}

func (s *service) deleteLimits(ctx context.Context, st Storage, ids []uint64, scopeIDs []uint64) error {
	scopedIDs, err := st.DeleteLimits(ctx, ids, scopeIDs)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			With("limits_ids", ids).
			Error("delete limits failed")

		return err
	}

	err = st.DeleteCounters(ctx, scopedIDs)
	if err != nil {
		s.logger.WithCtx(ctx).
			WithError(err).
			With("limits_ids", ids).
			Error("delete counters failed")

		return err
	}

	return nil
}

func splitLimits(limits []domain.Limit) ([]domain.Limit, []domain.Limit) {
	var (
		static   []domain.Limit
		periodic []domain.Limit
	)

	for _, limit := range limits {
		switch limit.LimitType {
		case domain.LimitTypeMINAMOUNT, domain.LimitTypeMAXAMOUNT:
			static = append(static, limit)
		case domain.LimitTypeTOTALAMOUNT, domain.LimitTypeTOTALCOUNT:
			periodic = append(periodic, limit)
		}
	}

	return static, periodic
}

func checkStaticLimits(limits []domain.Limit, amount decimal.Decimal) error {
	for _, limit := range limits {
		switch limit.LimitType {
		case domain.LimitTypeMINAMOUNT:
			if amount.LessThan(limit.Value) {
				return ctxerrors.New(
					ctxerrors.TypeInternal,
					fmt.Sprintf("operation amount %s is less than min_limit value %s", amount, limit.Value),
				)
			}
		case domain.LimitTypeMAXAMOUNT:
			if amount.GreaterThan(limit.Value) {
				return ctxerrors.New(
					ctxerrors.TypeInternal,
					fmt.Sprintf("operation amount %s is greater than max_limit value %s", amount, limit.Value),
				)
			}
		}
	}

	return nil
}

func (s *service) GenerateCounters(ctx context.Context, limits []domain.Limit, now time.Time) ([]domain.Counter, error) {
	counters := make([]domain.Counter, 0, len(limits))

	for _, limit := range limits {
		start, end, err := generateStartEndPeriods(limit, now)
		if err != nil {
			s.logger.WithCtx(ctx).
				WithError(err).
				With("limit_id", limit.ID).
				Error("generate start end periods failed")

			return nil, err
		}

		hash, err := GenerateCounterHash(limit, start)
		if err != nil {
			s.logger.WithCtx(ctx).
				WithError(err).
				With("limit_id", limit.ID).
				Error("generate counter hash failed")

			return nil, err
		}

		counter := domain.Counter{
			LimitID:   limit.ID,
			Hash:      hash,
			StartTime: start,
			EndTime:   end,
		}

		counters = append(counters, counter)
	}

	return counters, nil
}

func mergeEntities(target domain.Attributes, source domain.Attributes) (domain.Attributes, error) {
	attributes := make(map[string]domain.Attribute)

	for _, attribute := range target {
		attributes[attribute.Name] = attribute
	}

	for _, attribute := range source {
		current, ok := attributes[attribute.Name]
		if ok {
			if current.Value != attribute.Value {
				return nil, ctxerrors.New(
					ctxerrors.TypeInvalidRequest,
					fmt.Sprintf("attribute %s has different values %s and %s", attribute.Name, current.Value, attribute.Value),
				)
			}
		}
		attributes[attribute.Name] = attribute
	}

	target = maps.Values(attributes)
	domain.SortEntities(target)

	return target, nil
}

func entitiesIsEqual(a, b domain.Attributes) bool {
	return len(a) == len(b)
}

func filterOperations(operations []domain.Operation) ([]domain.Operation, error) {
	filtered := make([]domain.Operation, 0, len(operations))

	for _, operation := range operations {
		if operation.Status == domain.OperationStatusCommitted || operation.Status == domain.OperationStatusRollback {
			return nil, ctxerrors.New(
				ctxerrors.TypeInternal,
				"operation is not pending",
			)
		}
		if operation.Status == domain.OperationStatusPending {
			filtered = append(filtered, operation)
		}
	}

	return filtered, nil
}

func validateExistedOperations(operations []domain.Operation) (domain.Operation, error) {
	filtered, err := filterOperations(operations)
	if err != nil {
		return domain.Operation{}, err
	}

	if len(filtered) == 0 {
		return domain.Operation{}, ctxerrors.New(
			ctxerrors.TypeInvalidRequest,
			"operation is not pending",
		)
	}

	return filtered[0], nil
}

func filterOldLimits(new, old []domain.Limit) []domain.Limit {
	limits := make(map[uint64]domain.Limit)

	for _, limit := range new {
		limits[limit.ID] = limit
	}

	for _, limit := range old {
		delete(limits, limit.ID)
	}

	return maps.Values(limits)
}
