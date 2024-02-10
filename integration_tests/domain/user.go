package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type User struct {
	ID        int             `json:"id" db:"id"`
	Name      string          `json:"name" db:"name"`
	Balance   decimal.Decimal `json:"balance" db:"balance"`
	CratedAt  time.Time       `json:"-" db:"created_at"`
	UpdatedAt time.Time       `json:"-" db:"updated_at"`
}
