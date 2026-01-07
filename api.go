package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/milossilhar/boot-go-http-server/internal/auth"
	"github.com/milossilhar/boot-go-http-server/internal/database"
)

type apiConfig struct {
	environment    string
	jwtSecret      string
	queries        *database.Queries
	fileServerHits atomic.Int32
}

func (ac *apiConfig) metricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Hits: %d", ac.fileServerHits.Load())))
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (ac *apiConfig) postUserHandler(w http.ResponseWriter, r *http.Request) {
	type createUserDTO struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	decoded := createUserDTO{}
	err := decoder.Decode(&decoded)

	if err != nil {
		log.Printf("Error decoding JSON: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	hashedPassword, err := auth.HashPassword(decoded.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	user, err := ac.queries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          decoded.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		log.Printf("Error creating user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	type response struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}
	createUserResponse := response{
		ID:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Local().UTC().Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Local().UTC().Format(time.RFC3339),
	}
	respondWithJSON(w, http.StatusCreated, createUserResponse)
}

func (ac *apiConfig) putUserHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		log.Printf("User ID not found in context")
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	type updateUserDTO struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	decoded := updateUserDTO{}
	err := decoder.Decode(&decoded)
	if err != nil {
		log.Printf("Error decoding JSON: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	hashedPassword, err := auth.HashPassword(decoded.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	user, err := ac.queries.UpdateMyself(r.Context(), database.UpdateMyselfParams{
		Email:          decoded.Email,
		HashedPassword: hashedPassword,
		ID:             userId,
	})
	if err != nil {
		log.Printf("Error updating user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	type response struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}
	putUserResponse := response{
		ID:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Local().UTC().Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Local().UTC().Format(time.RFC3339),
	}
	respondWithJSON(w, http.StatusOK, putUserResponse)

}

func (ac *apiConfig) getAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	type userDTO struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	users, err := ac.queries.GetAllUsers(r.Context())
	if err != nil {
		log.Printf("Error getting all users: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response := make([]userDTO, len(users))
	for i, user := range users {
		response[i] = userDTO{
			ID:        user.ID.String(),
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
			UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		}
	}

	respondWithJSON(w, http.StatusOK, response)
}

func validateChirp(chirp string, w http.ResponseWriter) bool {
	if len(chirp) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return false
	}

	return true
}

func censorBadWords(chirp string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(chirp, " ")
	for i := 0; i < len(words); i++ {
		if slices.Contains(badWords, strings.ToLower(words[i])) {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}

func (ac *apiConfig) postChirpHandler(w http.ResponseWriter, r *http.Request) {
	type createChirpDTO struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	decoded := createChirpDTO{}
	err := decoder.Decode(&decoded)

	if err != nil {
		log.Printf("Error decoding JSON: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

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

	if !validateChirp(decoded.Body, w) {
		return
	}

	censoredChirp := censorBadWords(decoded.Body)

	chirp, err := ac.queries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   censoredChirp,
		UserID: userID,
	})
	if err != nil {
		log.Printf("Error creating chirp: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	respondWithJSON(w, http.StatusCreated, toChirpDTO(chirp))
}

func (ac *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		log.Printf("User ID not found in context")
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	chirpID, err := pathVariable(r, "chirpID", uuid.Parse)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Chirp ID is required")
		return
	}

	chirp, err := ac.queries.GetChirp(r.Context(), chirpID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Chirp not found")
		default:
			log.Printf("Error getting chirp: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		}
		return
	}

	if chirp.UserID != userId {
		respondWithError(w, http.StatusForbidden, "You can only delete your own chirps")
		return
	}

	_, err = ac.queries.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		log.Printf("Error deleting chirp: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (ac *apiConfig) getAllChirpsHandler(w http.ResponseWriter, r *http.Request) {
	chirps, err := ac.queries.GetAllChirps(r.Context())
	if err != nil {
		log.Printf("Error getting all chirps: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	response := make([]chirpDTO, len(chirps))
	for i, chirp := range chirps {
		response[i] = toChirpDTO(chirp)
	}

	respondWithJSON(w, http.StatusOK, response)
}

type chirpDTO struct {
	ID        string `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	UserID    string `json:"user_id"`
}

func toChirpDTO(ch database.Chirp) chirpDTO {
	return chirpDTO{
		ID:        ch.ID.String(),
		Body:      ch.Body,
		CreatedAt: ch.CreatedAt.Format(time.RFC3339),
		UpdatedAt: ch.UpdatedAt.Format(time.RFC3339),
		UserID:    ch.UserID.String(),
	}
}

func (ac *apiConfig) getChirpHandler(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	if chirpID == "" {
		respondWithError(w, http.StatusBadRequest, "Chirp ID is required")
		return
	}

	err := uuid.Validate(chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Chirp ID")
		return
	}

	parsedChirpID, _ := uuid.Parse(chirpID)
	chirp, err := ac.queries.GetChirp(r.Context(), parsedChirpID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Chirp not found")
		default:
			log.Printf("Error getting chirp: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, toChirpDTO(chirp))
}

func (ac *apiConfig) postLoginHandler(w http.ResponseWriter, r *http.Request) {
	type loginDTO struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	decoded := loginDTO{}
	err := decoder.Decode(&decoded)
	if err != nil {
		log.Printf("Error decoding JSON: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	user, err := ac.queries.GetUserByEmail(r.Context(), decoded.Email)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "User not found")
		default:
			log.Printf("Error getting user: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		}
		return
	}

	isValid, err := auth.VerifyPassword(decoded.Password, user.HashedPassword)
	if err != nil || !isValid {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	token, err := auth.MakeJWT(user.ID, ac.jwtSecret, "chirpy", 1*time.Second)
	if err != nil {
		log.Printf("Error creating JWT: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Error creating refresh token: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	_, err = ac.queries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
	})
	if err != nil {
		log.Printf("Error saving refresh token: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	type response struct {
		ID           string `json:"id"`
		Email        string `json:"email"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	loginResponse := response{
		ID:           user.ID.String(),
		Email:        user.Email,
		CreatedAt:    user.CreatedAt.Local().UTC().Format(time.RFC3339),
		UpdatedAt:    user.UpdatedAt.Local().UTC().Format(time.RFC3339),
		Token:        token,
		RefreshToken: refreshToken,
	}
	respondWithJSON(w, http.StatusOK, loginResponse)
}

func (ac *apiConfig) postRefreshHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting Bearer token: %v", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	refreshToken, err := ac.queries.GetRefreshToken(r.Context(), token)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error getting refresh token: %v", err)
		}
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	if refreshToken.RevokedAt.Valid || refreshToken.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	token, err = auth.MakeJWT(refreshToken.UserID, ac.jwtSecret, "chirpy", 1*time.Hour)
	if err != nil {
		log.Printf("Error creating JWT: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	type response struct {
		Token string `json:"token"`
	}
	respondWithJSON(w, http.StatusOK, response{
		Token: token,
	})
}

func (ac *apiConfig) postRevokeHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting Bearer token: %v", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	rows, err := ac.queries.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		log.Printf("Error revoking refresh token: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	if rows != 1 {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (ac *apiConfig) registerAPI() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /metrics", ac.metricsHandler())

	// auth
	mux.HandleFunc("POST /login", ac.postLoginHandler)
	mux.HandleFunc("POST /refresh", ac.postRefreshHandler)
	mux.HandleFunc("POST /revoke", ac.postRevokeHandler)

	// users
	mux.HandleFunc("GET /users", ac.getAllUsersHandler)
	mux.HandleFunc("POST /users", ac.postUserHandler)
	mux.HandleFunc("PUT /users", ac.middlewareAuthenticated(ac.putUserHandler))

	// chirps
	mux.HandleFunc("GET /chirps", ac.getAllChirpsHandler)
	mux.HandleFunc("GET /chirps/{chirpID}", ac.getChirpHandler)
	mux.HandleFunc("POST /chirps", ac.postChirpHandler)
	mux.HandleFunc("DELETE /chirps/{chirpID}", ac.middlewareAuthenticated(ac.deleteChirpHandler))

	return mux
}
