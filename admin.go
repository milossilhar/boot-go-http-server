package main

import (
	"fmt"
	"log"
	"net/http"
)

func (ac *apiConfig) adminMetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>GOLang Server ADMIN</title>
  </head>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
`, ac.fileServerHits.Load())))
	}
}

func (ac *apiConfig) adminResetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ac.environment != "development" {
			respondWithError(w, http.StatusForbidden, "Forbidden")
			return
		}

		ac.fileServerHits.Store(0)
		rows, err := ac.queries.DeleteAllUsers(r.Context())
		if err != nil {
			log.Printf("Error deleting all users: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Counter reset. Deleted %d user(s)", rows)))
	}
}

func (ac *apiConfig) registerADMIN() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /metrics", ac.adminMetricsHandler())
	mux.HandleFunc("POST /reset", ac.adminResetHandler())

	return mux
}
