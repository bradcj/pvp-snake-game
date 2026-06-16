package network

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bradcj/pvp-snake-game/internal/game"
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
	clientID := conn.RemoteAddr().String()
	defer func() {
		log.Printf("Client %s disconnected\n", clientID)
		hub.Unregister <- clientID // ensure client is unregistered on disconnect
		conn.Close()
	}()

	log.Printf("Client %s connected\n", clientID)
	hub.Register <- clientID // register client in the hub

	for {
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Client %s disconnected or error encountered while reading message: %v\n", clientID, err)
			break
		}

		log.Printf("Client %s sent message: %s of type %d\n", clientID, payload, messageType)
		hub.IncomingActions <- game.Action{ClientID: clientID, Data: string(payload)}
	}
}
