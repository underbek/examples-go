package handler

import (
	"net/http"
	"strconv"

	"layout/domain"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	user, err := parseRequest[domain.User](r)
	if err != nil {
		h.logger.Error("parse request", zap.Error(err))
		sendResponse[NilType](w, nil, err)
		return
	}

	user, err = h.useCase.CreateUser(r.Context(), user)
	if err != nil {
		h.logger.Error("CreateUser failed", zap.Error(err))
		sendResponse[NilType](w, nil, err)
		return
	}

	sendResponse(w, user, nil)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		h.logger.Error("parse user id", zap.Error(err))
		sendResponse[NilType](w, nil, err)
		return
	}

	user, err := h.useCase.GetUser(r.Context(), userID)
	if err != nil {
		h.logger.Error("GetUser failed", zap.Error(err))
		sendResponse[NilType](w, nil, err)
		return
	}

	sendResponse(w, user, nil)
}
