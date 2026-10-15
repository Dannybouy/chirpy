package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/dannybouy/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is not set")
	}
	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	dbQueries := database.New(dbConn)
	
	apiConfig := apiConfig{
		fileserverHits: atomic.Int32{},
		db: dbQueries,
	}
	fileServerHandler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))

	//Create a new http.ServeMux to route requests.
	mux := http.NewServeMux()

	// Register the app handler for the /app path and strip the /app prefix.
	mux.Handle("/app/", apiConfig.middlewareMetricsInc(fileServerHandler))

	// Register the healthz handler for the /healthz path.
	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	// Register the metrics handler for the /metrics path.
	mux.HandleFunc("GET /admin/metrics", apiConfig.handlerMetrics)

	// Register the reset metrics handler for the /reset path.
	mux.HandleFunc("POST /admin/reset", apiConfig.handlerResetMetrics)

	// Register validate_chirp endpoint
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	// Create a new http.Server struct and set the Address and Handler.
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
