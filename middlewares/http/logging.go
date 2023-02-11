package httpmiddleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/underbek/examples-go/buffer"
	"github.com/underbek/examples-go/logger"
)

func Logging(logger *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			start := time.Now()

			l := logger.
				WithCtx(r.Context()).
				With("method", r.Method).
				With("path", r.URL.Path).
				With("addr", r.RemoteAddr).
				With("user_agent", r.UserAgent())

			buf := buffer.NewMemoryBuffer(1024)

			_, err := io.Copy(buf, r.Body)
			if err != nil {
				l.Error(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			l.
				With("request_body", string(buf.Bytes())).
				Debug("got request")

			r.Body = io.NopCloser(bytes.NewBuffer(buf.Bytes()))

			rec := httptest.NewRecorder()

			next.ServeHTTP(rec, r)

			_, err = w.Write(rec.Body.Bytes())
			if err != nil {
				l.Error(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			for key, values := range rec.Header() {
				for _, value := range values {
					w.Header().Add(key, value)
				}
			}

			l.
				With("response", rec.Body.String()).
				With("duration", time.Since(start)).
				Debug("response sent")
		})
	}
}