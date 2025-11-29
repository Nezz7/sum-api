package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
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
	log.Println("Received request:", r.Method, r.URL.Path, r.URL.RawQuery)

	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		if err := json.NewEncoder(w).Encode(SumResponse{Error: "Method not allowed"}); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
		}
		return
	}

	aStr := r.URL.Query().Get("a")
	bStr := r.URL.Query().Get("b")

	if aStr == "" || bStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(SumResponse{Error: "Parameters 'a' and 'b' are required"}); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
		}
		return
	}

	a, err := strconv.ParseInt(aStr, 10, 64)
	if err != nil {
		log.Println("Error parsing 'a':", err)
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(SumResponse{Error: "Invalid parameter 'a'"}); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
		}
		return
	}

	b, err := strconv.ParseInt(bStr, 10, 64)
	if err != nil {
		log.Println("Error parsing 'b':", err)
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(SumResponse{Error: "Invalid parameter 'b'"}); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
		}
		if err := json.NewEncoder(w).Encode(SumResponse{Error: "Invalid parameter 'b'"}); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
		}
		return
	}

	result, err := Add(a, b)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(SumResponse{Error: err.Error()}); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
		}
		return
	}

	response := SumResponse{Result: result}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

func main() {
	API_KEY := "tXPT/PA+1jbHCCQEvuQ6sfMGB9rx7S/T72TRufZZADI="
	log.Println(API_KEY)
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

	log.Printf("Sum-API server is starting on %s\n", addr)
	log.Fatal(server.ListenAndServe())
}
