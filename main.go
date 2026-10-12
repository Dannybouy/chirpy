package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

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
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(fmt.Appendf([]byte{}, "Hits: %d\n", cfg.countRequestHits()))
}

func main() {
	const filepathRoot = "."
	const port = "8080"
	apiConfig := apiConfig{
		fileserverHits: atomic.Int32{},
	}
	fileServerHandler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))

	//Create a new http.ServeMux to route requests.
	mux := http.NewServeMux()

	// Register the app handler for the /app path and strip the /app prefix.
	mux.Handle("/app/", apiConfig.middlewareMetricsInc(fileServerHandler))

	// Register the healthz handler for the /healthz path.
	mux.HandleFunc("GET /healthz", handlerReadiness)

	// Register the metrics handler for the /metrics path.
	mux.HandleFunc("GET /metrics", apiConfig.handlerMetrics)

	// Register the reset metrics handler for the /reset path.
	mux.HandleFunc("POST /reset", apiConfig.handlerResetMetrics)

	// Create a new http.Server struct and set the Address and Handler.
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
