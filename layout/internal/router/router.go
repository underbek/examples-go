package router

import (
	"net/http"

	"layout/internal/handler"

	"github.com/gorilla/mux"
)

func New(h *handler.Handler) http.Handler {
	r := mux.NewRouter()
	r.Use(SetJSONHeader)

	r.HandleFunc("/healthz", h.HealthCheck).Methods(http.MethodGet)

	apiRouter := r.PathPrefix("/api/v1").Subrouter()
	apiRouter.Use(SetJSONHeader)

	apiRouter.HandleFunc("/users/{id}", h.GetUser).Methods(http.MethodGet)
	apiRouter.HandleFunc("/users", h.CreateUser).Methods(http.MethodPost)

	apiRouter.HandleFunc("/orders/{id}", h.GetOrder).Methods(http.MethodGet)
	apiRouter.HandleFunc("/orders", h.CreateOrder).Methods(http.MethodPost)

	return r
}

func SetJSONHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
