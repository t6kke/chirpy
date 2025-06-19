package main

import (
	"log"
	"time"
	"net/http"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/t6kke/chirpy/internal/database"
	"github.com/t6kke/chirpy/internal/auth"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	ChirpyRed    bool      `json:"is_chirpy_red"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func (cfg *apiConfig) handlerAddUser(w http.ResponseWriter, r *http.Request) {
	type RequestedUser struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	new_user_req := RequestedUser{}
	err := decoder.Decode(&new_user_req)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	hashed_password, err := auth.HashPassword(new_user_req.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		w.WriteHeader(500)
		return
	}

	create_user_parameters := database.CreateUserParams{
		Email:          new_user_req.Email,
		HashedPassword: hashed_password,
	}

	db_user, err := cfg.dbq.CreateUser(r.Context(), create_user_parameters)
	if err != nil {
		log.Printf("Error creating user: %s", err)
		w.WriteHeader(500) //TODO need better response to return info that user already exists
		return
	}

	response_user := User{
		ID:        db_user.ID,
		CreatedAt: db_user.CreatedAt,
		UpdatedAt: db_user.UpdatedAt,
		Email:     db_user.Email,
		ChirpyRed: db_user.IsChirpyRed,
	}
	response_data, err := json.Marshal(response_user)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response_data)
}

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	type RequestedUser struct {
		Password     string        `json:"password"`
		Email        string        `json:"email"`
		ExpiresInSec time.Duration `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	new_user_req := RequestedUser{}
	err := decoder.Decode(&new_user_req)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	if new_user_req.ExpiresInSec <=0 || new_user_req.ExpiresInSec >= 3600 {
		new_user_req.ExpiresInSec = 3600 * time.Second
	} else {
		new_user_req.ExpiresInSec = new_user_req.ExpiresInSec * time.Second
	}

	db_user, err := cfg.dbq.FindUserWithEmail(r.Context(), new_user_req.Email)
	if err != nil {
		log.Printf("Error creating user: %s", err)
		w.WriteHeader(500) //TODO need better response to return info that user already exists
		return
	}

	err = auth.CheckPasswordHash(db_user.HashedPassword, new_user_req.Password)
	if err != nil {
		log.Printf("failed to log in: %s", err)
		w.WriteHeader(401)
		w.Write([]byte("Incorrect email or password"))
		return
	}

	token, err := auth.MakeJWT(db_user.ID, cfg.c_secret, new_user_req.ExpiresInSec)
	if err != nil {
		log.Printf("Failed to generate token: %s", err)
		w.WriteHeader(500)
		return
	}
	refresh_token, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Failed to generate refresh token: %s", err)
		w.WriteHeader(500)
		return
	}

	r_token_insert_param := database.CreateRefreshTokenParams{
		Token:     refresh_token,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
		UserID:    db_user.ID,
	}

	_, err = cfg.dbq.CreateRefreshToken(r.Context(), r_token_insert_param)
	if err != nil {
		log.Printf("Failed to insert refresh token to DB: %s", err)
		w.WriteHeader(500)
		return
	}

	response_user := User{
		ID:           db_user.ID,
		CreatedAt:    db_user.CreatedAt,
		UpdatedAt:    db_user.UpdatedAt,
		Email:        db_user.Email,
		ChirpyRed: db_user.IsChirpyRed,
		Token:        token,
		RefreshToken: refresh_token,
	}
	response_data, err := json.Marshal(response_user)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response_data)
}

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	refresh_token_from_header, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Failed to extract refresh token from header: %s", err)
		w.WriteHeader(401)
		w.Write([]byte("Failed to extract refresh token from header"))
		return
	}

	db_user_from_r_token, err := cfg.dbq.GetUserFromRefreshToken(r.Context(), refresh_token_from_header)
	if err != nil {
		log.Printf("No valid refresh token found: %s", err)
		w.WriteHeader(401)
		w.Write([]byte("No valid refresh token found"))
		return
	}

	new_JWT, err := auth.MakeJWT(db_user_from_r_token, cfg.c_secret, 3600 * time.Second)
	if err != nil {
		log.Printf("Failed to generate token: %s", err)
		w.WriteHeader(500)
		return
	}

	type returntoken struct {
		Token string `json:"token"`
	}
	return_token := returntoken{
		Token: new_JWT,
	}

	response_data, err := json.Marshal(return_token)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response_data)
}

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	refresh_token_from_header, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Failed to extract refresh token from header: %s", err)
		w.WriteHeader(401)
		w.Write([]byte("Failed to extract refresh token from header"))
		return
	}

	_, err = cfg.dbq.RevokeRefreshToken(r.Context(), refresh_token_from_header)
	if err != nil {
		log.Printf("No valid refresh token found: %s", err)
		w.WriteHeader(401)
		w.Write([]byte("No valid refresh token found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerUpdateUserPwEm(w http.ResponseWriter, r *http.Request) {
	token_from_header, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Failed to extract token from header: %s", err)
		w.WriteHeader(401)
		w.Write([]byte("Failed to extract token from header"))
		return
	}
	user_id_from_token, err := auth.ValidateJWT(token_from_header, cfg.c_secret)
	if err != nil {
		log.Printf("Token mismatch: %s", err)
		w.WriteHeader(401)
		w.Write([]byte("Invalid Token"))
		return
	}

	//TODO maybe before proceeding we need to check if user currently also exists in DB

	type RequestedUpdate struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	requested_update := RequestedUpdate{}
	err = decoder.Decode(&requested_update)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}
	hashed_password, err := auth.HashPassword(requested_update.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		w.WriteHeader(500)
		return
	}

	update_parameters := database.UpdatePasswordAndEmailParams{
		ID:             user_id_from_token,
		Email:          requested_update.Email,
		HashedPassword: hashed_password,
	}

	db_user, err := cfg.dbq.UpdatePasswordAndEmail(r.Context(), update_parameters)
	if err != nil {
		log.Printf("Error updateing user email and password: %s", err)
		w.WriteHeader(500)
		return
	}

	response_user := User{
		ID:        db_user.ID,
		CreatedAt: db_user.CreatedAt,
		UpdatedAt: db_user.UpdatedAt,
		Email:     db_user.Email,
		ChirpyRed: db_user.IsChirpyRed,
	}
	response_data, err := json.Marshal(response_user)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response_data)
}
