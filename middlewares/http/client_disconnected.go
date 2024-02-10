package httpmiddleware

import (
	"context"
	"errors"
	"net/http"
)

func ClientDisconnectMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		next.ServeHTTP(w, r)
		if errors.Is(ctx.Err(), context.Canceled) {
			w.WriteHeader(499)
		}
	})
}
