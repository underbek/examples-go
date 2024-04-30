package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

/*
ENUM(
MIN_AMOUNT
MAX_AMOUNT
TOTAL_AMOUNT
TOTAL_COUNT
)
*/
type LimitType int

/*
ENUM(
CALENDAR_DAY
CALENDAR_WEEK
CALENDAR_MONTH
)
*/
type PeriodType int

type Limit struct {
	ID        uint64          `json:"id" db:"id" updateApi:"limit_id"`
	Hash      string          `json:"hash" db:"hash"`
	LimitType LimitType       `json:"limit_type" db:"limit_type" updateApi:"limit_type"`
	Currency  string          `json:"currency" db:"currency" updateApi:"currency"`
	Value     decimal.Decimal `json:"value" db:"value" updateApi:"value"`
	Entities  Attributes      `json:"entities" db:"meta" updateApi:"entities"`
	Period    *PeriodType     `json:"period,omitempty" db:"period" updateApi:"period"`
	Timezone  *string         `json:"timezone,omitempty" db:"timezone" updateApi:"timezone"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}
