package main

import (
	"log"
	"net/http"
	"ws/internal/handlers"

	"github.com/bmizerany/pat"
)

// routes defines the application routes
func routes() http.Handler {
	log.Println("Entered in routes")
	mux := pat.New()

	mux.Get("/", http.HandlerFunc(handlers.Home))
	mux.Get("/ws", http.HandlerFunc(handlers.WsEndpoint))

	fileServer := http.FileServer(http.Dir("./static/"))
	log.Println("set fileserver")
	mux.Get("/static/", http.StripPrefix("/static", fileServer))

	return mux
}
