package handlers

import (
	"encoding/json"
	"net/http"
)

type balanceHandler struct {
	logic logicInt
}

func NewBalanceHandler(logic logicInt) http.Handler {
	return &balanceHandler{
		logic: logic,
	}
}

func (h *balanceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var request UpdateBalanceRequest
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&request); err != nil {
			createErrorResponse(err, w)
			return
		}

		user, err := h.logic.AddBalance(r.Context(), request.UserId, request.Amount)
		if err != nil {
			createErrorResponse(err, w)
			return
		}

		createResponse(UserResponse{user}, http.StatusOK, w)
		return
	}
}
