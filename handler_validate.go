package main

import (
	"encoding/json"
	"net/http"
	"regexp"
)

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	// Dealing with the request body (parsing the JSON/Decoding the request body)
	type parameters struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}
	if params.Body == "" {
		respondWithError(w, http.StatusBadRequest, "Chirp is empty", nil)
		return
	}
	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	sentence := params.Body

	replacementWords := map[string]string {
		"kerfuffle": "****",
		"sharbert": "****",
		"fornax": "****",
	}

	for oldWord, newWord := range replacementWords {
		// (?i) makes the regex case-insensitive
		// \b matches a word boundary

		pattern := `(?i)\b` + regexp.QuoteMeta(oldWord) + `\b`
		re := regexp.MustCompile(pattern)
		sentence = re.ReplaceAllString(sentence, newWord)
	}

	// Encoding the response body (encoding the response body)
	type returnVals struct {
		Cleaned_Body string `json:"cleaned_body"`
	}
	respBody := returnVals {
		Cleaned_Body: sentence,
	}
	respondWithJSON(w, http.StatusOK, respBody)
}