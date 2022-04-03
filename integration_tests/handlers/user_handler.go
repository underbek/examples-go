package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/AndreyAndreevich/examples-go/integration_tests/domain"
)

//go:generate mockery --name=logicInt --structname=logicIntMock --filename=logicint_mock.go --inpackage
type logicInt interface {
	CreateUser(name string, balance float64) (domain.User, error)
	GetUser(id int) (domain.User, error)
	AddBalance(id int, amount float64) (domain.User, error)
}

type userHandler struct {
	logic logicInt
}

func NewUserHandler(logic logicInt) http.Handler {
	return &userHandler{
		logic: logic,
	}
}

func (h *userHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var request CreateUserRequest
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&request); err != nil {
			createErrorResponse(err, w)
			return
		}

		user, err := h.logic.CreateUser(request.Name, request.Balance)
		if err != nil {
			createErrorResponse(err, w)
			return
		}

		createResponse(UserResponse{user}, http.StatusOK, w)
		return

	case http.MethodGet:
		var request GetUserRequest
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&request); err != nil {
			createErrorResponse(err, w)
			return
		}

		user, err := h.logic.GetUser(request.Id)
		if err != nil {
			createErrorResponse(err, w)
			return
		}

		createResponse(UserResponse{user}, http.StatusOK, w)
		return
	}
}

func createErrorResponse(err error, w http.ResponseWriter) {
	createResponse(
		ErrorResponse{
			Error: err.Error(),
		},
		http.StatusInternalServerError,
		w,
	)
}

func createResponse(data interface{}, status int, w http.ResponseWriter) {
	encoder := json.NewEncoder(w)
	fmt.Println(encoder.Encode(data))
	w.WriteHeader(status)
}
