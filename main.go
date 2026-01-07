package main

import (
	"database/sql"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/milossilhar/boot-go-http-server/internal/database"
)

func main() {
	godotenv.Load()

	dbURL := envRequired("DB_URL")
	port := env("PORT", "8080")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()

	serverConfig := apiConfig{
		environment:    env("ENVIRONMENT", "development"),
		jwtSecret:      envRequired("JWT_SECRET"),
		polkaKey:       envRequired("POLKA_KEY"),
		queries:        database.New(db),
		fileServerHits: atomic.Int32{},
	}

	mux.Handle("/app/", http.StripPrefix("/app", serverConfig.registerAPP()))
	mux.Handle("/api/", http.StripPrefix("/api", serverConfig.registerAPI()))
	mux.Handle("/admin/", http.StripPrefix("/admin", serverConfig.registerADMIN()))

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Println("Listening on port " + port)
	log.Fatal(server.ListenAndServe())
}
