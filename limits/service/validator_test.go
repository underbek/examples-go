package service

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/underbek/examples-go/limits/domain"
	"github.com/underbek/examples-go/utils"
)

func TestValidateLimit(t *testing.T) {
	tests := []struct {
		name       string
		limit      *domain.Limit
		isError    bool
		errMessage string
	}{
		{
			name: "value less than zero",
			limit: &domain.Limit{
				LimitType: domain.LimitTypeMAXAMOUNT,
				Value:     decimal.NewFromInt(-1),
				Currency:  "EUR",
				Entities:  []domain.Attribute{{Name: "merchant_id", Value: "1"}},
			},
			isError:    true,
			errMessage: "limit value is less than or equal zero, but value = -1",
		},
		{
			name: "value equal zero",
			limit: &domain.Limit{
				LimitType: domain.LimitTypeMAXAMOUNT,
				Value:     decimal.Zero,
				Currency:  "EUR",
				Entities:  []domain.Attribute{{Name: "merchant_id", Value: "1"}},
			},
			isError:    true,
			errMessage: "limit value is less than or equal zero, but value = 0",
		},
		{
			name: "empty currency",
			limit: &domain.Limit{
				LimitType: domain.LimitTypeMAXAMOUNT,
				Value:     decimal.NewFromInt(100),
				Entities:  []domain.Attribute{{Name: "merchant_id", Value: "1"}},
			},
			isError:    true,
			errMessage: "limit currency is empty",
		},
		{
			name: "empty entities",
			limit: &domain.Limit{
				LimitType: domain.LimitTypeMAXAMOUNT,
				Value:     decimal.NewFromInt(100),
				Currency:  "EUR",
			},
			isError:    true,
			errMessage: "entities is empty",
		},
		{
			name: "incorrect entity name",
			limit: &domain.Limit{
				LimitType: domain.LimitTypeMAXAMOUNT,
				Value:     decimal.NewFromInt(100),
				Currency:  "EUR",
				Entities: []domain.Attribute{{
					Name:  "Incorrect",
					Value: "100",
				}},
			},
			isError:    true,
			errMessage: "entity name has an upper case symbol: \"Incorrect\"",
		},
		{
			name: "empty entity value",
			limit: &domain.Limit{
				LimitType: domain.LimitTypeMAXAMOUNT,
				Value:     decimal.NewFromInt(100),
				Currency:  "EUR",
				Entities: []domain.Attribute{{
					Name: "merchant_id",
				}},
			},
			isError:    true,
			errMessage: "entity value with name \"merchant_id\" is empty",
		},
		{
			name: "duplicated entity names",
			limit: &domain.Limit{
				LimitType: domain.LimitTypeMAXAMOUNT,
				Value:     decimal.NewFromInt(100),
				Currency:  "EUR",
				Entities: []domain.Attribute{
					{
						Name:  "merchant_id",
						Value: "1",
					},
					{
						Name:  "merchant_id",
						Value: "2",
					},
				},
			},
			isError:    true,
			errMessage: "entity with name \"merchant_id\" is duplicated",
		},
		{
			name: "amount limit has a period",
			limit: &domain.Limit{
				LimitType: domain.LimitTypeMINAMOUNT,
				Value:     decimal.NewFromInt(100),
				Currency:  "EUR",
				Entities:  []domain.Attribute{{Name: "merchant_id", Value: "1"}},
				Period:    utils.ToPtr(domain.PeriodTypeCALENDARDAY),
			},
			isError:    true,
			errMessage: "limit with min_amount type has a period",
		},
		{
			name: "amount limit has a timezone",
			limit: &domain.Limit{
				LimitType: domain.LimitTypeMAXAMOUNT,
				Value:     decimal.NewFromInt(100),
				Currency:  "EUR",
				Entities:  []domain.Attribute{{Name: "merchant_id", Value: "1"}},
				Timezone:  utils.ToPtr("UTC"),
			},
			isError:    true,
			errMessage: "limit with max_amount type has a timezone",
		},
		{
			name: "total limit doesn't have a period",
			limit: &domain.Limit{
				LimitType: domain.LimitTypeTOTALCOUNT,
				Value:     decimal.NewFromInt(100),
				Currency:  "EUR",
				Entities:  []domain.Attribute{{Name: "merchant_id", Value: "1"}},
				Timezone:  utils.ToPtr("UTC"),
			},
			isError:    true,
			errMessage: "limit with total_count type doesn't have a period",
		},
		{
			name: "total limit doesn't have a timezone",
			limit: &domain.Limit{
				LimitType: domain.LimitTypeTOTALAMOUNT,
				Value:     decimal.NewFromInt(100),
				Currency:  "EUR",
				Entities:  []domain.Attribute{{Name: "merchant_id", Value: "1"}},
				Period:    utils.ToPtr(domain.PeriodTypeCALENDARDAY),
			},
			isError: false,
		},
		{
			name: "invalid limit value",
			limit: &domain.Limit{
				LimitType: domain.LimitTypeTOTALCOUNT,
				Value:     decimal.NewFromFloat(100.02),
				Currency:  "EUR",
				Entities:  []domain.Attribute{{Name: "merchant_id", Value: "1"}},
				Period:    utils.ToPtr(domain.PeriodTypeCALENDARDAY),
				Timezone:  utils.ToPtr("UTC"),
			},
			isError:    true,
			errMessage: "is not integer",
		},
		{
			name: "invalid timezone",
			limit: &domain.Limit{
				LimitType: domain.LimitTypeTOTALAMOUNT,
				Value:     decimal.NewFromInt(100),
				Currency:  "EUR",
				Entities:  []domain.Attribute{{Name: "merchant_id", Value: "1"}},
				Period:    utils.ToPtr(domain.PeriodTypeCALENDARDAY),
				Timezone:  utils.ToPtr("Invalid"),
			},
			isError:    true,
			errMessage: "invalid timezone",
		},
		{
			name: "validated amount limit",
			limit: &domain.Limit{
				LimitType: domain.LimitTypeMAXAMOUNT,
				Value:     decimal.NewFromInt(100),
				Currency:  "EUR",
				Entities:  []domain.Attribute{{Name: "merchant_id", Value: "1"}},
			},
		},
		{
			name: "validated total limit",
			limit: &domain.Limit{
				LimitType: domain.LimitTypeTOTALCOUNT,
				Value:     decimal.NewFromInt(100),
				Currency:  "EUR",
				Entities:  []domain.Attribute{{Name: "merchant_id", Value: "1"}},
				Period:    utils.ToPtr(domain.PeriodTypeCALENDARDAY),
				Timezone:  utils.ToPtr("Europe/London"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLimit(tt.limit)
			if tt.isError {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.errMessage)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestValidateOperationInfo(t *testing.T) {
	tests := []struct {
		name       string
		operation  domain.OperationInfo
		isError    bool
		errMessage string
	}{
		{
			name: "value less than zero",
			operation: domain.OperationInfo{
				Amount: domain.Amount{
					Value:    decimal.NewFromInt(-1),
					Currency: "EUR",
				},
				Meta: []domain.Attribute{{Name: "merchant_id", Value: "1"}},
			},
			isError:    true,
			errMessage: "operation amount is less than or equal zero, but value = -1",
		},
		{
			name: "value equal zero",
			operation: domain.OperationInfo{
				Amount: domain.Amount{
					Value:    decimal.NewFromInt(0),
					Currency: "EUR",
				},
				Meta: []domain.Attribute{{Name: "merchant_id", Value: "1"}},
			},
			isError:    true,
			errMessage: "operation amount is less than or equal zero, but value = 0",
		},
		{
			name: "empty currency",
			operation: domain.OperationInfo{
				Amount: domain.Amount{
					Value: decimal.NewFromInt(100),
				},
				Meta: []domain.Attribute{{Name: "merchant_id", Value: "1"}},
			},
			isError:    true,
			errMessage: "operation currency is empty",
		},
		{
			name: "empty entities",
			operation: domain.OperationInfo{
				Amount: domain.Amount{
					Value:    decimal.NewFromInt(100),
					Currency: "EUR",
				},
			},
			isError:    true,
			errMessage: "entities is empty",
		},
		{
			name: "incorrect entity name",
			operation: domain.OperationInfo{
				Amount: domain.Amount{
					Value:    decimal.NewFromInt(100),
					Currency: "EUR",
				},
				Meta: []domain.Attribute{{Name: "Incorrect", Value: "1"}},
			},
			isError:    true,
			errMessage: "entity name has an upper case symbol: \"Incorrect\"",
		},
		{
			name: "empty entity value",
			operation: domain.OperationInfo{
				Amount: domain.Amount{
					Value:    decimal.NewFromInt(100),
					Currency: "EUR",
				},
				Meta: []domain.Attribute{{Name: "merchant_id"}},
			},
			isError:    true,
			errMessage: "entity value with name \"merchant_id\" is empty",
		},
		{
			name: "duplicated entity names",
			operation: domain.OperationInfo{
				Amount: domain.Amount{
					Value:    decimal.NewFromInt(100),
					Currency: "EUR",
				},
				Meta: []domain.Attribute{
					{
						Name:  "merchant_id",
						Value: "1",
					},
					{
						Name:  "merchant_id",
						Value: "2",
					},
				},
			},
			isError:    true,
			errMessage: "entity with name \"merchant_id\" is duplicated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOperationInfo(tt.operation)
			if tt.isError {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.errMessage)
				return
			}

			require.NoError(t, err)
		})
	}
}
