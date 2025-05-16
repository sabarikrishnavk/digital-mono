package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RequestDurationTimer is an interface for timing request durations.
type RequestDurationTimer interface {
	ObserveDuration() // Records the elapsed time since the timer was started.
}

// PrometheusMetrics defines an interface for our metrics system.
type PrometheusMetrics interface {
	IncRequestsTotal(operation, handlerType string)
	IncResponsesTotal(operation, handlerType, statusCode string) // Simplified based on seller_handler usage
	NewRequestDurationTimer(operation, handlerType string) RequestDurationTimer
	Handler() http.Handler // To expose metrics endpoint
	// Add other metrics methods like IncErrors, ObserveRequestDuration, etc.
}

// promMetrics is a concrete implementation using Prometheus (currently placeholder).
type promMetrics struct {
	requests *prometheus.CounterVec // Example
	requestsTotal   *prometheus.CounterVec
	responsesTotal  *prometheus.CounterVec // Added to store the responsesTotal metric
	requestDuration *prometheus.HistogramVec // To track request durations
	serviceName     string
	subsystem       string
}

// NewPrometheusMetrics creates a new PrometheusMetrics instance.
func NewPrometheusMetrics(serviceName, subsystem string) PrometheusMetrics {
	// In a real scenario, you would initialize and register Prometheus collectors here.
	// Example:
	requests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: serviceName,
			Subsystem: subsystem,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{"path", "method", "code"},
	)
	// Note: The seller_handler uses operation/handlerType, not path/method/code for IncRequestsTotal.
	// Let's adjust the metric definition to match the handler's usage for simplicity.
	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: serviceName,
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "Total number of requests by operation and handler type.",
		},
		[]string{"operation", "handler_type"},
	)

	responsesTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: serviceName,
			Subsystem: subsystem,
			Name:      "responses_total",
			Help:      "Total number of responses by operation, handler type, and status code.",
		},
		[]string{"operation", "handler_type", "code"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: serviceName,
			Subsystem: subsystem,
			Name:      "request_duration_seconds",
			Help:      "Request duration in seconds by operation and handler type.",
			Buckets:   prometheus.DefBuckets, // Default buckets (0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10)
		},
		[]string{"operation", "handler_type"},
	)

	prometheus.MustRegister(requests,requestsTotal, responsesTotal, requestDuration)

	fmt.Printf("Metrics: Initialized for service %s, subsystem %s\n", serviceName, subsystem)
	return &promMetrics{
		requests : requests,
		requestsTotal:   requestsTotal,
		responsesTotal:  responsesTotal, // Assign the created metric
		requestDuration: requestDuration,
		serviceName:     serviceName,
		subsystem:       subsystem,
	}
}

// Handler returns an http.Handler for exposing Prometheus metrics.
func (pm *promMetrics) Handler() http.Handler {
	return promhttp.Handler() // Uncomment when using actual prometheus client
	// return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprintln(w, "# Placeholder metrics endpoint")
	// })
}

// IncRequestsTotal increments a request counter by operation and handler type.
func (pm *promMetrics) IncRequestsTotal(operation, handlerType string) {
	pm.requestsTotal.WithLabelValues(operation, handlerType).Inc()
}

// IncResponsesTotal increments a response counter by operation, handler type, and status code.
func (pm *promMetrics) IncResponsesTotal(operation, handlerType, statusCode string) {
	pm.responsesTotal.WithLabelValues(operation, handlerType, statusCode).Inc()
}

// promTimer is a concrete implementation of RequestDurationTimer.
type promTimer struct {
	start time.Time
	observer prometheus.Observer // Can be a Histogram or Summary
}

// ObserveDuration records the elapsed time.
func (pt *promTimer) ObserveDuration() {
	duration := time.Since(pt.start).Seconds()
	fmt.Printf("Metrics: Observing duration %.4f seconds\n", duration)
	pt.observer.Observe(duration)
}

// NewRequestDurationTimer starts a timer for a given operation and handler type.
func (pm *promMetrics) NewRequestDurationTimer(operation, handlerType string) RequestDurationTimer {
	fmt.Printf("Metrics (%s - %s): Starting timer for operation: %s, handler_type: %s\n", pm.serviceName, pm.subsystem, operation, handlerType)
	observer := pm.requestDuration.WithLabelValues(operation, handlerType)
	return &promTimer{
		start: time.Now(),
		observer: observer,
	}
}