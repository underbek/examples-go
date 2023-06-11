package httpmiddleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/underbek/examples-go/buffer"
	"github.com/underbek/examples-go/logger"
)

// HealthCheckPath defines default health check path to services
const HealthCheckPath = "/health_check"

func Logging(logger *logger.Logger, showHealthLogs bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//if requested path contains health_check path, and we need to hide
			//health check logs, then return
			missLogger := false
			if strings.Contains(r.URL.Path, HealthCheckPath) && !showHealthLogs {
				missLogger = true
			}

			start := time.Now()

			l := logger.
				WithCtx(r.Context()).
				With("method", r.Method).
				With("path", r.URL.Path).
				With("addr", r.RemoteAddr).
				With("user_agent", r.UserAgent())

			buf := buffer.NewMemoryBuffer()

			_, err := io.Copy(buf, r.Body)
			if err != nil {
				if !missLogger {
					l.Error(err.Error())
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if !missLogger {
				l.With("request_body", string(buf.Bytes())).
					Debug("got request")
			}

			r.Body = io.NopCloser(bytes.NewBuffer(buf.Bytes()))

			rec := newWriter(w)
			next.ServeHTTP(rec, r)

			for key, values := range rec.Header() {
				for _, value := range values {
					w.Header().Add(key, value)
				}
			}

			if !missLogger {
				l.With("code", rec.StatusCode()).
					With("response", rec.Body()).
					With("duration", time.Since(start)).
					Debug("response sent")
			}
		})
	}
}
