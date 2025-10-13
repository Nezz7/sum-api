package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

type SumResponse struct {
	Result int64  `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func fib(n int) int {
	if n == 0 {
		return 0
	}
	if n == 1 {
		return 1
	}
	return fib(n-1) + fib(n-2)
}

func OptimizedFib(n int) int {
	if n == 0 {
		return 0
	}
	a := 0
	b := 1
	for i := 2; i <= n; i++ {
		tmp := a + b
		a = b
		b = tmp
	}
	return b
}

func Add(a, b int64) (int64, error) {
	if a > 0 && b > 0 && a > math.MaxInt64-b {
		return 0, fmt.Errorf("integer overflow: %d + %d exceeds maximum value", a, b)
	}
	if a < 0 && b < 0 && a < math.MinInt64-b {
		return 0, fmt.Errorf("integer overflow: %d + %d exceeds minimum value", a, b)
	}
	return a + b, nil
}

func sumHandler(w http.ResponseWriter, r *http.Request) {
	// Create logger with request context
	logger := slog.With(
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)

	logger.Info("request received")

	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		logger.Warn("method not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(SumResponse{Error: "Method not allowed"})
		return
	}

	aStr := r.URL.Query().Get("a")
	bStr := r.URL.Query().Get("b")

	if aStr == "" || bStr == "" {
		logger.Warn("missing parameters")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SumResponse{Error: "Parameters 'a' and 'b' are required"})
		return
	}

	a, err := strconv.ParseInt(aStr, 10, 64)
	if err != nil {
		logger.Warn("invalid parameter",
			slog.String("param", "a"),
			slog.String("value", aStr),
		)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SumResponse{Error: "Invalid parameter 'a'"})
		return
	}

	b, err := strconv.ParseInt(bStr, 10, 64)
	if err != nil {
		logger.Warn("invalid parameter",
			slog.String("param", "b"),
			slog.String("value", bStr),
		)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SumResponse{Error: "Invalid parameter 'b'"})
		return
	}

	logger = logger.With(
		slog.Int64("a", a),
		slog.Int64("b", b),
	)

	result, err := Add(a, b)
	if err != nil {
		logger.Warn("operation failed",
			slog.String("error", err.Error()),
		)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SumResponse{Error: err.Error()})
		return
	}

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
