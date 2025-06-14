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
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
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

	response_user := User{
		ID:        db_user.ID,
		CreatedAt: db_user.CreatedAt,
		UpdatedAt: db_user.UpdatedAt,
		Email:     db_user.Email,
		Token:     token,
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
