package main

import (
	"log"
	"time"
	"net/http"
	"encoding/json"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerAddUser(w http.ResponseWriter, r *http.Request) {
	type newRequestedUserEmail struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	new_user_email := newRequestedUserEmail{}
	err := decoder.Decode(&new_user_email)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	db_user, err := cfg.dbq.CreateUser(r.Context(), new_user_email.Email)
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
