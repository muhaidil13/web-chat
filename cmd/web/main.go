package main

import (
	"log"
	"net/http"

	"github.com/web-chat/internal/handlers"
)

func main() {
	router := route()
	log.Println("Starting Channel listener")
	go handlers.ListenToWsChannel()

	log.Println("Starting Web Server on Port 8080")
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	err := srv.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}
