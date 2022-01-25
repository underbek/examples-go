package handlers

import "github.com/AndreyAndreevich/examples-go/integration_tests/domain"

type CreateUserRequest struct {
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
}

type GetUserRequest struct {
	Id int `json:"id"`
}

type UpdateBalanceRequest struct {
	UserId int     `json:"user_id"`
	Amount float64 `json:"amount"`
}

type UserResponse struct {
	domain.User
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func foo() {
	print("Hi")
}
