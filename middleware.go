package main

import (
	"context"
	"log"
	"net/http"

	"github.com/milossilhar/boot-go-http-server/internal/auth"
)

func (ac *apiConfig) middlewareIncreaseHits(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		ac.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (ac *apiConfig) middlewareAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			log.Printf("Error getting Bearer token: %v", err)
			respondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}
		userID, err := auth.ValidateJWT(token, ac.jwtSecret, "chirpy")
		if err != nil {
			log.Printf("Error validating JWT: %v", err)
			respondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		authenticatedR := r.WithContext(context.WithValue(r.Context(), "userID", userID))

		next(w, authenticatedR)
	}
}
