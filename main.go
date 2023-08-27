package main

import (
	"log"
	"net/http"
	"ws/internal/handlers"
)

func main() {
	mux := routes()

	log.Println("Starting channel listener")
	go handlers.ListenToWsChannel()

	log.Println("Staring web server on port 8080")

	err := http.ListenAndServe("0.0.0.0:8080", mux)

	if err != nil {
		log.Printf("error in ListenAndServe: %v \n", err)
	}
}
