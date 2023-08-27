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

	port := os.Getenv("PORT")
	if port == ""{
		port = "3000"
	}
	log.Println("Staring web server on port:" + port)

	err := http.ListenAndServe("0.0.0.0:"+port, mux)

	if err != nil {
		log.Printf("error in ListenAndServe: %v \n", err)
	}
}
