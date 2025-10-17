package main

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddSimple(t *testing.T) {
	result, err := Add(context.Background(), 2, 3)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != 5 {
		t.Errorf("Expected 5, got %d", result)
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int64
		expected int64
		hasError bool
	}{
		{"positive numbers", 5, 3, 8, false},
		{"negative numbers", -5, -3, -8, false},
		{"mixed numbers", -5, 3, -2, false},
		{"zero values", 0, 0, 0, false},
		{"max int64", math.MaxInt64, 0, math.MaxInt64, false},
		{"min int64", math.MinInt64, 0, math.MinInt64, false},
		{"positive overflow", math.MaxInt64, 1, 0, true},
		{"negative overflow", math.MinInt64, -1, 0, true},
		{"large positive overflow", math.MaxInt64, math.MaxInt64, 0, true},
		{"large negative overflow", math.MinInt64, math.MinInt64, 0, true},
		{"edge case positive", math.MaxInt64 - 1, 1, math.MaxInt64, false},
		{"edge case negative", math.MinInt64 + 1, -1, math.MinInt64, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Add(context.Background(), tt.a, tt.b)
			if tt.hasError {
				if err == nil {
					t.Errorf("Add(%d, %d) expected error but got none", tt.a, tt.b)
				}
			} else {
				if err != nil {
					t.Errorf("Add(%d, %d) unexpected error: %v", tt.a, tt.b, err)
				}
				if result != tt.expected {
					t.Errorf("Add(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
				}
			}
		})
	}
}

func TestSumHandler_GET(t *testing.T) {
	tests := []struct {
		name           string
		a, b           string
		expectedStatus int
		expectedResult int64
		expectError    bool
	}{
		{"valid sum", "5", "3", http.StatusOK, 8, false},
		{"negative sum", "-10", "5", http.StatusOK, -5, false},
		{"overflow case", "9223372036854775807", "1", http.StatusBadRequest, 0, true},
		{"large numbers", "1000000", "2000000", http.StatusOK, 3000000, false},
		{"missing parameter a", "", "5", http.StatusBadRequest, 0, true},
		{"missing parameter b", "5", "", http.StatusBadRequest, 0, true},
		{"invalid parameter a", "abc", "5", http.StatusBadRequest, 0, true},
		{"invalid parameter b", "5", "xyz", http.StatusBadRequest, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/sum"
			if tt.a != "" || tt.b != "" {
				url += "?a=" + tt.a + "&b=" + tt.b
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)

			w := httptest.NewRecorder()
			sumHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response SumResponse
			json.NewDecoder(w.Body).Decode(&response)

			if tt.expectError {
				if response.Error == "" {
					t.Error("Expected error but got none")
				}
				if response.Result != 0 {
					t.Errorf("Expected no result field when error, got %d", response.Result)
				}
			} else {
				if response.Error != "" {
					t.Errorf("Unexpected error: %s", response.Error)
				}
				if response.Result != tt.expectedResult {
					t.Errorf("Expected result %d, got %d", tt.expectedResult, response.Result)
				}
			}
		})
	}
}

func TestSumHandler_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/sum", nil)
	w := httptest.NewRecorder()

	sumHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}
}

func BenchmarkAdd(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		Add(ctx, 123456789, 987654321)
	}
}

func BenchmarkAddLargeNumbers(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		Add(ctx, math.MaxInt64-1000, 500)
	}
}

func BenchmarkAddOverflow(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		Add(ctx, math.MaxInt64, 1)
	}
}

func BenchmarkFib(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fib(35)
	}
}

func BenchmarkOptimizedFib(b *testing.B) {
	for i := 0; i < b.N; i++ {
		OptimizedFib(35)
	}
}
