package main

import (
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/web-chat/internal/handlers"
)

func route() http.Handler {
	mux := pat.New()
	mux.Get("/", http.HandlerFunc(handlers.Home))
	mux.Get("/ws", http.HandlerFunc(handlers.WsEndPoint))
	filesever := http.FileServer(http.Dir("../../static"))
	mux.Get("/static/", http.StripPrefix("/static", filesever))
	return mux
}
