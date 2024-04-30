package httpmiddleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/underbek/examples-go/logger"
)

func MuxRecoveryMiddleware(h http.Handler, logger *logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.WithCtx(r.Context()).
					With("method", r.Method).
					With("url", r.URL.String()).
					With("trace", string(debug.Stack())).
					With("panic", err).
					Error(fmt.Sprintf("Recovered from pgtaskpool panic: %v", err))
			}
		}()

		h.ServeHTTP(w, r)
	})
}
