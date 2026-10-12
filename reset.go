package main

import (
	"net/http"
)

// handlerResetMetrics resets the fileserverHits counter to zero
func (cfg *apiConfig) handlerResetMetrics(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to zero"))
}