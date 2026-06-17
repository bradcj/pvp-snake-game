package network

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bradcj/pvp-snake-game/internal/game"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins for dev
	},
}

// upgrade HTTP connection to WebSocket connection
func HandleWebSocket(hub *game.CentralHub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Error upgrading to WebSocket: %v\n", err)
		return
	}
	client := &game.Client{
		ID:       uuid.New().String(),
		Conn:     conn,
		SendChan: make(chan []byte, 256),
	}

	defer func() {
		log.Printf("Client %s disconnected\n", client.ID)
		hub.Unregister <- *client // ensure client is unregistered on disconnect
		conn.Close()
	}()

	log.Printf("Client %s connected\n", client.ID)
	hub.Register <- *client // register client in the hub

	// seperate goroutines for reading and writing to the WebSocket connection
	go WritePump(client, hub)
	ReadPump(client, hub) // this will block until the client disconnects or an error occurs
}
