package main

import (
	"log"
	"net/http"
	"os"
	"ws/internal/handlers"
)

func main() {
	mux := routes()

	log.Println("Starting channel listener")
	go handlers.ListenToWsChannel()

	log.Println("Staring web server on port 8080")

	port := os.Getenv("PORT")
	if port == ""{
		port = "3000"
	}
	err := http.ListenAndServe("0.0.0.0:%d"+port, mux)

	if err != nil {
		log.Printf("error in ListenAndServe: %v \n", err)
	}
}
