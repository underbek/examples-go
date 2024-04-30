package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

/*
ENUM(
new
pending
committed
rollback
)
*/
type OperationStatus int

type Operation struct {
	ID        uint64          `json:"id" db:"id"`
	ContextID uint64          `json:"context_id" db:"context_id"`
	Currency  string          `json:"currency" db:"currency"`
	Value     decimal.Decimal `json:"value" db:"value"`
	Status    OperationStatus `json:"status" db:"status"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}

type OperationInfo struct {
	Amount Amount     `json:"amount"`
	Meta   Attributes `json:"entities"`
}

type AppendOperationInfo struct {
	ContextID uint64     `json:"context_id"`
	Meta      Attributes `json:"entities"`
}

type FinalizeOperationsInfo struct {
	ContextID uint64          `json:"context_id"`
	Status    OperationStatus `json:"status"`
}
