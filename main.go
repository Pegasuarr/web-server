package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Response is a standard JSON response structure
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Logger middleware — logs method, path, status, and duration
func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[%s] %s %s — %v", r.Method, r.RemoteAddr, r.URL.Path, time.Since(start))
		next.ServeHTTP(w, r)
	})
}

// writeJSON sends a JSON response with the given status code
func writeJSON(w http.ResponseWriter, status int, payload Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Message: "only POST method is allowed",
		})
		return
	}

	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("failed to parse form: %v", err),
		})
		return
	}

	name := r.FormValue("name")
	address := r.FormValue("address")

	if name == "" || address == "" {
		writeJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Message: "name and address are required",
		})
		return
	}

	log.Printf("Form submitted — name: %s, address: %s", name, address)

	writeJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "form submitted successfully",
		Data: map[string]string{
			"name":    name,
			"address": address,
		},
	})
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/hello" {
		writeJSON(w, http.StatusNotFound, Response{
			Success: false,
			Message: "404 — page not found",
		})
		return
	}

	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Message: "only GET method is allowed",
		})
		return
	}

	writeJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "Hello, World!",
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "server is running",
		Data: map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		},
	})
}

func main() {
	logger := log.New(os.Stdout, "[SERVER] ", log.LstdFlags)

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fileServer)
	mux.HandleFunc("/form", formHandler)
	mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("/health", healthHandler)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	logger.Println("Server starting on http://localhost:8080")
	logger.Println("Routes: / | /form | /hello | /health")

	if err := server.ListenAndServe(); err != nil {
		logger.Fatalf("Server failed to start: %v", err)
	}
}
