package httpmiddleware

import "net/http"

type logResponseWriter struct {
	statusCode int
	body       []byte
	internal   http.ResponseWriter
}

func newWriter(w http.ResponseWriter) *logResponseWriter {
	return &logResponseWriter{
		internal: w,
	}
}

func (w *logResponseWriter) Header() http.Header {
	return w.internal.Header()
}

func (w *logResponseWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return w.internal.Write(w.body)
}

func (w *logResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.internal.WriteHeader(statusCode)
}

func (w *logResponseWriter) StatusCode() int {
	return w.statusCode
}

func (w *logResponseWriter) Body() string {
	return string(w.body)
}
