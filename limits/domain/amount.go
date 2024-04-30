package domain

import "github.com/shopspring/decimal"

type Amount struct {
	Currency string          `json:"currency"`
	Value    decimal.Decimal `json:"value"`
}
