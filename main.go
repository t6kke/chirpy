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
	server_mux.Handle("/", file_server)

	server_struct := &http.Server{
		Addr:    ":"+ port,
		Handler: server_mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server_struct.ListenAndServe())
}
