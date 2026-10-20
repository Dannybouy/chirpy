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
	db             *database.Queries
	platform       string
	jwtSecret      string
	polkaKey       string
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	jwtSecret := os.Getenv("JWT_SECRET")
	polkaKey := os.Getenv("POLKA_KEY")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set")
	}
	if dbURL == "" {
		log.Fatal("DB_URL is not set")
	}
	if polkaKey == "" {
		log.Fatal("POLKA_KEY is not set")
	}
	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	dbQueries := database.New(dbConn)

	apiConfig := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		jwtSecret:      jwtSecret,
		polkaKey:       polkaKey,
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

	// Register the refresh token endpoint
	mux.HandleFunc("POST /api/refresh", apiConfig.handlerRefreshToken)

	// Register a new revoke token endpoint
	mux.HandleFunc("POST /api/revoke", apiConfig.handlerRevokeToken)

	// Register the create user endpoint
	mux.HandleFunc("POST /api/users", apiConfig.handlerCreateUser)

	// Register the update user endpoint
	mux.HandleFunc("PUT /api/users", apiConfig.handlerUpdateUser)

	// Register the login endpoint
	mux.HandleFunc("POST /api/login", apiConfig.handlerLoginUser)

	// Register a new chirp endpoint
	mux.HandleFunc("POST /api/chirps", apiConfig.handlerCreateChirp)

	// Register a new Get All Chirps endpoint
	mux.HandleFunc("GET /api/chirps", apiConfig.handlerGetAllChirps)

	// Register a new Get Chirp endpoint
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiConfig.handlerGetSingleChirp)

	// Register a new delete chirp endpoint
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiConfig.handlerDeleteChirp)

	// Register the polka webhook endpoint
	mux.HandleFunc("POST /api/polka/webhooks", apiConfig.handlerPolkaWebhooks)

	// Create a new http.Server struct and set the Address and Handler.
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
