package httpmiddleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/underbek/examples-go/metrics"
)

func MuxMetricsMiddleware() func(http.Handler) http.Handler {

	collector := newMuxCollector()
	prometheus.MustRegister(collector)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			route := mux.CurrentRoute(r)
			path, _ := route.GetPathTemplate()
			rec := newWriter(w)

			next.ServeHTTP(rec, r)

			collector.requestCounter.WithLabelValues(path, fmt.Sprintf("%d", rec.StatusCode())).Inc()
			collector.requestDurationHistogram.WithLabelValues(path).Observe(time.Since(start).Seconds())
		})
	}
}

type httpMuxCollector struct {
	requestCounter           *prometheus.CounterVec
	requestDurationHistogram *prometheus.HistogramVec
}

func newMuxCollector() *httpMuxCollector {
	return &httpMuxCollector{
		requestCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "http",
				Subsystem: "mux",
				Name:      "request_total",
				Help:      "Total number of requests.",
			},
			[]string{"path", "code"},
		),
		requestDurationHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "http",
				Subsystem: "mux",
				Name:      "request_duration_seconds",
				Help:      "Histogram of request duration (seconds).",
				Buckets:   metrics.DefBuckets,
			},
			[]string{"path"},
		),
	}
}

func (c httpMuxCollector) Describe(descs chan<- *prometheus.Desc) {
	c.requestCounter.Describe(descs)
	c.requestDurationHistogram.Describe(descs)
}

func (c httpMuxCollector) Collect(metrics chan<- prometheus.Metric) {
	c.requestCounter.Collect(metrics)
	c.requestDurationHistogram.Collect(metrics)
}
