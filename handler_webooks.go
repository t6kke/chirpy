package main

import (
	"log"
	"net/http"
	"encoding/json"

	"github.com/google/uuid"

	//"github.com/t6kke/chirpy/internal/database"
)

func (cfg *apiConfig) handlerPolkaPaymentUpgrade(w http.ResponseWriter, r *http.Request) {
	type InputDataStruct struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	input_data := InputDataStruct{}
	err := decoder.Decode(&input_data)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	if input_data.Event != "user.upgraded" {
		log.Printf("Event type '%s' did not match 'user.upgraded'", input_data.Event)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	user_id, err := uuid.Parse(input_data.Data.UserID)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	_, err = cfg.dbq.UpgradeUserChirpyRed(r.Context(), user_id)
	if err != nil {
		log.Printf("Error upgradeing user ChirpyRed status: %s", err)
		w.WriteHeader(404)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
