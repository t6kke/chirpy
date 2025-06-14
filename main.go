package main

import (
	"os"
	"log"
	"net/http"
	"sync/atomic"
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"

	"github.com/t6kke/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbq            *database.Queries
	platform       string
	c_secret       string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}
	chirpy_secret := os.Getenv("CHIRPY_SECRET")
	if chirpy_secret == "" {
		log.Fatal("CHIRPY_SECRET must be set")
	}
	const filepathRoot = "."
	const port = "8080"

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()
	dbQueries := database.New(db)

	var counter atomic.Int32
	api_cfg := apiConfig{
		fileserverHits: counter,
		dbq:            dbQueries,
		platform:       platform,
		c_secret:       chirpy_secret,
	}

	server_mux := http.NewServeMux()
	file_server := http.FileServer(http.Dir(filepathRoot))
	server_mux.Handle("/app/", api_cfg.middlewareMetricsInc(http.StripPrefix("/app", file_server)))
	server_mux.HandleFunc("GET /api/healthz", handlerReadiness)
	server_mux.HandleFunc("GET /api/chirps", api_cfg.handlerGetAllChirps)
	server_mux.HandleFunc("GET /api/chirps/{chirpID}", api_cfg.handlerGetOneChirp)
	server_mux.HandleFunc("POST /api/chirps", api_cfg.handlerAddChirp)
	server_mux.HandleFunc("POST /api/users", api_cfg.handlerAddUser)
	server_mux.HandleFunc("POST /api/login", api_cfg.handlerUserLogin)
	server_mux.HandleFunc("GET /admin/metrics", api_cfg.handlerMetrics)
	server_mux.HandleFunc("POST /admin/reset", api_cfg.handlerReset)

	server_struct := &http.Server{
		Addr:    ":"+ port,
		Handler: server_mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server_struct.ListenAndServe())
}


