package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type Counter struct {
	Hash      string    `json:"hash" db:"hash"`
	LimitID   uint64    `json:"limit_id" db:"limit_id"`
	StartTime time.Time `json:"start_time" db:"start_time"`
	EndTime   time.Time `json:"end_time" db:"end_time"`
}

type ExceededCounters struct {
	CounterID  uint64          `json:"counter_id" db:"counter_id"`
	LimitID    uint64          `json:"limit_id" db:"limit_id"`
	LimitType  string          `json:"limit_type" db:"limit_type"`
	Period     PeriodType      `json:"period" db:"period"`
	Meta       Attributes      `json:"meta" db:"meta"`
	LimitValue decimal.Decimal `json:"limit_value" db:"limit_value"`
	NewValue   decimal.Decimal `json:"new_value" db:"new_value"`
}
