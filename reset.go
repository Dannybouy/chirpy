package main

import (
	"net/http"
	"errors"
)

// handlerResetMetrics resets the fileserverHits counter to zero
func (cfg *apiConfig) handlerResetMetrics(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Not allowed", errors.New("not allowed"))
		return
	}
	err := cfg.db.DeleteUser(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't reset metrics", err)
		return
	}
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to zero"))
}