package domain

import "github.com/shopspring/decimal"

type Order struct {
	ID     int64           `json:"id" db:"id"`
	Amount decimal.Decimal `json:"amount" db:"amount"`
}
