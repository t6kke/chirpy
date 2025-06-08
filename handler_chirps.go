package main

import (
	"log"
	"time"
	"strings"
	"net/http"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/t6kke/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	db_all_chirps, err := cfg.dbq.GetAllChirps(r.Context())
	if err != nil {
		log.Printf("Error creating user: %s", err)
		w.WriteHeader(500) //TODO need better response to return info that failed to retreive all chirps
		return
	}

	result_slice := make([]Chirp, 0)

	for _, db_chirp := range db_all_chirps {
		response_chirp := Chirp{
			ID:        db_chirp.ID,
			CreatedAt: db_chirp.CreatedAt,
			UpdatedAt: db_chirp.UpdatedAt,
			Body:      db_chirp.Body,
			UserID:    db_chirp.UserID,
		}
		result_slice = append(result_slice, response_chirp)
	}


	response_data, err := json.Marshal(result_slice)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response_data)
}


func (cfg *apiConfig) handlerAddChirp(w http.ResponseWriter, r *http.Request) {
	type chirp_body struct {
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	type error_return struct {
		Error string `json:"error"`
	}

	decoder := json.NewDecoder(r.Body)
	c_body := chirp_body{}
	err := decoder.Decode(&c_body)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	if len(c_body.Body) > 140 {
		error_response_body := error_return{
			Error: "Chirp is too long",
		}
		response_data, err := json.Marshal(error_response_body)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response_data)
		return
	}

	new_c_body := checkProfane(c_body.Body)

	query_insert_parameters := database.CreateChirpParams{
		Body:   new_c_body,
		UserID: c_body.UserID,
	}

	db_chirp, err := cfg.dbq.CreateChirp(r.Context(), query_insert_parameters)
	if err != nil {
		log.Printf("Error creating user: %s", err)
		w.WriteHeader(500) //TODO need better response to return info that failed to add chirp
		return
	}

	response_chirp := Chirp{
		ID:        db_chirp.ID,
		CreatedAt: db_chirp.CreatedAt,
		UpdatedAt: db_chirp.UpdatedAt,
		Body:      db_chirp.Body,
		UserID:    db_chirp.UserID,
	}
	response_data, err := json.Marshal(response_chirp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response_data)
}

func checkProfane(input string) string {
	bad_words := []string{"kerfuffle", "sharbert", "fornax"}

	c_words := strings.Split(input, " ")

	for i, word := range c_words {
		for _, b_word := range bad_words {
			if strings.ToLower(word) == b_word {
				c_words[i] = "****"
			}
		}
	}

	result_string := strings.Join(c_words, " ")
	return result_string
}
