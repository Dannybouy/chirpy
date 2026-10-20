package main

import (
	"fmt"
	"net/http"
)

// countRequestHits returns the number of requests served by the file server
func (cfg *apiConfig) countRequestHits() int32 {
	return cfg.fileserverHits.Load()
}

// middlewareMetricsInc increments the fileserverHits counter and serves the next handler
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// handlerMetrics returns the number of requests served by the file server
func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.countRequestHits())
}
