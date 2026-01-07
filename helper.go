package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
)

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, message)))
}

func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error encoding JSON: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write([]byte(response))
}

func pathVariable[T any](r *http.Request, name string, conv func(string) (T, error)) (T, error) {
	value := r.PathValue(name)
	if value == "" {
		var zero T
		return zero, errors.New("missing path variable " + name)
	}

	return conv(value)
}

func env(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func envRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatal("Environment variable " + key + " is required")
	}
	return value
}
