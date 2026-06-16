package main

import (
	"log"
	"net/http"

	"github.com/bradcj/pvp-snake-game/internal/game"
	"github.com/bradcj/pvp-snake-game/internal/network"
)

func main() {

	// spin up game loop in the background
	hub := game.NewHub()
	go hub.Run()

	// handle WebSocket connections
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		network.HandleWebSocket(hub, w, r)
	})

	// server static files for the client
	http.Handle("/", http.FileServer(http.Dir("./public")))

	log.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
