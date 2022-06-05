package handlers

import (
	"github.com/AndreyAndreevich/examples-go/integration_tests/domain"
	"github.com/shopspring/decimal"
)

type CreateUserRequest struct {
	Name    string          `json:"name"`
	Balance decimal.Decimal `json:"balance"`
}

type GetUserRequest struct {
	Id int `json:"id"`
}

type UpdateBalanceRequest struct {
	UserId int             `json:"user_id"`
	Amount decimal.Decimal `json:"amount"`
}

type UserResponse struct {
	domain.User
}

type ErrorResponse struct {
	Error string `json:"error"`
}
