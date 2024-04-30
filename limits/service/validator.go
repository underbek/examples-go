package service

import (
	"fmt"
	"time"
	"unicode"

	"github.com/shopspring/decimal"
	ctxerrors "github.com/underbek/examples-go/errors"
	"github.com/underbek/examples-go/limits/domain"
	"github.com/underbek/examples-go/utils"
)

const defaultTimeZone = "UTC"

func validateEntities(entities []domain.Attribute) error {
	const entitiesName = "entities"

	if len(entities) == 0 {
		return ctxerrors.New(
			ctxerrors.TypeInvalidRequest,
			"entities is empty",
		)
	}

	keys := make(map[string]struct{})
	for _, entity := range entities {
		for _, r := range entity.Name {
			if unicode.IsUpper(r) {
				return ctxerrors.New(
					ctxerrors.TypeInvalidRequest,
					fmt.Sprintf("entity name has an upper case symbol: \"%s\"", entity.Name),
				)
			}
		}

		if entity.Value == "" {
			return ctxerrors.New(
				ctxerrors.TypeInvalidRequest,
				fmt.Sprintf("entity value with name \"%s\" is empty", entity.Name),
			)
		}

		if _, ok := keys[entity.Name]; ok {
			return ctxerrors.New(
				ctxerrors.TypeInvalidRequest,
				fmt.Sprintf("entity with name \"%s\" is duplicated", entity.Name),
			)
		}
		keys[entity.Name] = struct{}{}
	}

	return nil
}

func validateLimit(limit *domain.Limit) error {
	if limit == nil {
		return ctxerrors.New(
			ctxerrors.TypeInvalidRequest,
			"limit is nil",
		)
	}

	if limit.Value.LessThanOrEqual(decimal.Zero) {
		return ctxerrors.New(
			ctxerrors.TypeInvalidRequest,
			fmt.Sprintf("limit value is less than or equal zero, but value = %s", limit.Value),
		)
	}

	if limit.Currency == "" {
		return ctxerrors.New(
			ctxerrors.TypeInvalidRequest,
			"limit currency is empty",
		)
	}

	err := validateEntities(limit.Entities)
	if err != nil {
		return err
	}

	switch limit.LimitType {
	case domain.LimitTypeMINAMOUNT, domain.LimitTypeMAXAMOUNT:
		if limit.Period != nil {
			return ctxerrors.New(
				ctxerrors.TypeInvalidRequest,
				fmt.Sprintf("limit with %s type has a period", limit.LimitType),
			)
		}

		if limit.Timezone != nil {
			return ctxerrors.New(
				ctxerrors.TypeInvalidRequest,
				fmt.Sprintf("limit with %s type has a timezone", limit.LimitType),
			)
		}
	case domain.LimitTypeTOTALAMOUNT, domain.LimitTypeTOTALCOUNT:
		if limit.Period == nil {
			return ctxerrors.New(
				ctxerrors.TypeInvalidRequest,
				fmt.Sprintf("limit with %s type doesn't have a period", limit.LimitType),
			)
		}

		if limit.Timezone == nil {
			limit.Timezone = utils.ToPtr(defaultTimeZone)
		}
	}

	if limit.LimitType == domain.LimitTypeTOTALCOUNT {
		if !limit.Value.IsInteger() {
			return ctxerrors.New(
				ctxerrors.TypeInvalidRequest,
				fmt.Sprintf("limit value %s is not integer", limit.Value),
			)
		}
	}

	if limit.Timezone != nil {
		_, err := time.LoadLocation(*limit.Timezone)
		if err != nil {
			return ctxerrors.Wrap(
				err,
				ctxerrors.TypeInvalidRequest,
				"invalid timezone",
			)
		}
	}

	return nil
}

func validateOperationInfo(operation domain.OperationInfo) error {
	if operation.Amount.Value.LessThanOrEqual(decimal.Zero) {
		return ctxerrors.New(
			ctxerrors.TypeInvalidRequest,
			fmt.Sprintf("operation amount is less than or equal zero, but value = %s", operation.Amount.Value),
		)
	}

	if operation.Amount.Currency == "" {
		return ctxerrors.New(
			ctxerrors.TypeInvalidRequest,
			"operation currency is empty",
		)
	}

	err := validateEntities(operation.Meta)
	if err != nil {
		return err
	}

	return nil
}
