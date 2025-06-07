package main

import (
	"log"
	"strings"
	"net/http"
	"encoding/json"
)

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type chirp_body struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	c_body := chirp_body{}
	err := decoder.Decode(&c_body)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	type error_return struct {
		Error string `json:"error"`
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

	type success_return struct {
		Cleaned_Body string `json:"cleaned_body"`
	}

	success_response_body := success_return{
		Cleaned_Body: new_c_body,
	}
	response_data, err := json.Marshal(success_response_body)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
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
