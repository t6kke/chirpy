package main

import (
	"log"
	"net/http"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	server_mux := http.NewServeMux()
	file_server := http.FileServer(http.Dir(filepathRoot))
	server_mux.Handle("/app/", http.StripPrefix("/app", file_server))
	server_mux.HandleFunc("/healthz", healthzHandler)

	server_struct := &http.Server{
		Addr:    ":"+ port,
		Handler: server_mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server_struct.ListenAndServe())
}


func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
