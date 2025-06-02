package main

import (
	"log"
	"net/http"
)

func main() {
	const port = "8080"
	server_mux := http.NewServeMux()

	server_struct := &http.Server{
		Addr:    ":"+ port,
		Handler: server_mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server_struct.ListenAndServe())
}
