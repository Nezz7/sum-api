package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type SumResponse struct {
	Result int64  `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

var (
	registry = prometheus.NewRegistry()

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

func init() {
	registry.MustRegister(httpRequestsTotal)
	registry.MustRegister(httpRequestDuration)
}

func Add(ctx context.Context, a, b int64) (int64, error) {
	tracer := otel.Tracer("sum-api")
	_, span := tracer.Start(ctx, "Add")
	defer span.End()

	if a > 0 && b > 0 && a > math.MaxInt64-b {
		return 0, fmt.Errorf("integer overflow: %d + %d exceeds maximum value", a, b)
	}
	if a < 0 && b < 0 && a < math.MinInt64-b {
		return 0, fmt.Errorf("integer overflow: %d + %d exceeds minimum value", a, b)
	}
	return a + b, nil
}

func sumHandler(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("sum-api")
	ctx, span := tracer.Start(r.Context(), "sumHandler")

	// Start timer for request duration
	start := time.Now()
	statusCode := http.StatusOK
	defer func() {
		httpRequestDuration.WithLabelValues(r.Method, "/sum").Observe(time.Since(start).Seconds())
		httpRequestsTotal.WithLabelValues(r.Method, "/sum", strconv.Itoa(statusCode)).Inc()
		span.SetAttributes(attribute.Int("http.status_code", statusCode))
		span.End()
	}()

	sleepDuration := time.Duration(rand.Float64() * 2 * float64(time.Second))
	time.Sleep(sleepDuration)

	// Create logger with request context
	logger := slog.With(
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("sleep", sleepDuration.String()),
		slog.String("trace_id", span.SpanContext().TraceID().String()),
	)

	logger.Info("request received")
	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.path", r.URL.Path),
	)

	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		logger.Warn("method not allowed")
		statusCode = http.StatusMethodNotAllowed
		span.SetStatus(codes.Error, "Method not allowed")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(SumResponse{Error: "Method not allowed"})
		return
	}

	aStr := r.URL.Query().Get("a")
	bStr := r.URL.Query().Get("b")

	if aStr == "" || bStr == "" {
		logger.Warn("missing parameters")
		statusCode = http.StatusBadRequest
		span.SetStatus(codes.Error, "Missing parameters")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(SumResponse{Error: "Parameters 'a' and 'b' are required"})
		return
	}

	a, err := strconv.ParseInt(aStr, 10, 64)
	if err != nil {
		logger.Warn("invalid parameter",
			slog.String("param", "a"),
			slog.String("value", aStr),
		)
		statusCode = http.StatusBadRequest
		span.SetStatus(codes.Error, "Invalid parameter 'a'")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(SumResponse{Error: "Invalid parameter 'a'"})
		return
	}

	b, err := strconv.ParseInt(bStr, 10, 64)
	if err != nil {
		logger.Warn("invalid parameter",
			slog.String("param", "b"),
			slog.String("value", bStr),
		)
		statusCode = http.StatusBadRequest
		span.SetStatus(codes.Error, "Invalid parameter 'b'")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(SumResponse{Error: "Invalid parameter 'b'"})
		return
	}

	logger = logger.With(
		slog.Int64("a", a),
		slog.Int64("b", b),
	)
	span.SetAttributes(
		attribute.Int64("sum.a", a),
		attribute.Int64("sum.b", b),
	)

	result, err := Add(ctx, a, b)
	if err != nil {
		logger.Warn("operation failed",
			slog.String("error", err.Error()),
		)
		statusCode = http.StatusBadRequest
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(SumResponse{Error: err.Error()})
		return
	}

	span.SetAttributes(attribute.Int64("sum.result", result))
	span.SetStatus(codes.Ok, "Success")
	logger.Info("sum calculated",
		slog.Int64("result", result),
	)
	json.NewEncoder(w).Encode(SumResponse{Result: result})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	// Initialize tracing
	shutdown, err := InitTracer()
	if err != nil {
		slog.Error("failed to initialize tracer", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown tracer", slog.String("error", err.Error()))
		}
	}()

	// Get log level from environment variable (default: INFO)
	logLevel := slog.LevelInfo
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		switch level {
		case "DEBUG":
			logLevel = slog.LevelDebug
		case "INFO":
			logLevel = slog.LevelInfo
		case "WARN":
			logLevel = slog.LevelWarn
		case "ERROR":
			logLevel = slog.LevelError
		}
	}
	// Use text format for local development, JSON for production
	var handler slog.Handler
	if os.Getenv("ENVIRONMENT") == "local" {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	http.HandleFunc("/sum", sumHandler)
	http.HandleFunc("/health", healthHandler)
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	addr := ":8080"
	server := &http.Server{
		Addr:         addr,
		Handler:      nil,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	slog.Info("server starting",
		slog.String("address", addr),
	)
	if err := server.ListenAndServe(); err != nil {
		slog.Error("server failed",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
}
