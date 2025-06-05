package main

import (
	//"fmt"
	"strconv"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerPrintMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits: " +  strconv.Itoa(int(cfg.fileserverHits.Load()))))
}

func (cfg *apiConfig) handlerResetMetrics(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	var counter atomic.Int32
	api_cfg := apiConfig{
		fileserverHits: counter,
	}

	server_mux := http.NewServeMux()
	file_server := http.FileServer(http.Dir(filepathRoot))
	server_mux.Handle("/app/", api_cfg.middlewareMetricsInc(http.StripPrefix("/app", file_server)))
	server_mux.HandleFunc("GET /healthz", healthzHandler)
	server_mux.HandleFunc("GET /metrics", api_cfg.handlerPrintMetrics)
	server_mux.HandleFunc("POST /reset", api_cfg.handlerResetMetrics)

	server_struct := &http.Server{
		Addr:    ":"+ port,
		Handler: server_mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server_struct.ListenAndServe())
}


