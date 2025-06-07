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
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
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
	}

	server_mux := http.NewServeMux()
	file_server := http.FileServer(http.Dir(filepathRoot))
	server_mux.Handle("/app/", api_cfg.middlewareMetricsInc(http.StripPrefix("/app", file_server)))
	server_mux.HandleFunc("GET  /api/healthz", handlerReadiness)
	server_mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)
	server_mux.HandleFunc("GET  /admin/metrics", api_cfg.handlerMetrics)
	server_mux.HandleFunc("POST /admin/reset", api_cfg.handlerResetMetrics)

	server_struct := &http.Server{
		Addr:    ":"+ port,
		Handler: server_mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server_struct.ListenAndServe())
}


