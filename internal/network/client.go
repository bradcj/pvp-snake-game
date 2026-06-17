package network

import (
	"encoding/json"
	"log"

	"github.com/bradcj/pvp-snake-game/internal/game"
	"github.com/gorilla/websocket"
)

const MAXIMUM_DROPPED_MESSAGES_IN_ROW = 10

// server -> client
func WritePump(client *game.Client, hub *game.CentralHub) {
	droppedMessages := 0
	for message := range client.SendChan {
		// break if client no longer part of hub
		_, exists := hub.Clients[client.ID]
		if !exists || client.Conn == nil {
			log.Printf("Client %s no longer connected, stopping WritePump\n", client.ID)
			break
		}
		err := client.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Error writing message to client %s: %v\n", client.ID, err)
			droppedMessages++
			if droppedMessages > MAXIMUM_DROPPED_MESSAGES_IN_ROW {
				log.Printf("Too many dropped messages for client %s, unregistering client\n", client.ID)
				hub.Unregister <- *client
				break
			}
			continue
		}
		droppedMessages = 0 // reset dropped message count on successful send
	}
}

// client -> server
func ReadPump(client *game.Client, hub *game.CentralHub) {
	defer func() {
		log.Printf("Client %s disconnected\n", client.ID)
		hub.Unregister <- *client // ensure client is unregistered on disconnect
		client.Conn.Close()
		client.Conn = nil // set connection to nil to signal WritePump to stop
	}()

	for {
		messageType, data, err := client.Conn.ReadMessage()
		if err != nil {
			log.Printf("Client %s disconnected or error encountered while reading message: %v\n", client.ID, err)
			hub.Unregister <- *client // ensure client is unregistered on disconnect
			break
		}
		log.Printf("Client %s sent message: %s of type %d\n", client.ID, data, messageType)
		if messageType != websocket.TextMessage {
			log.Printf("Client %s sent non-text message, ignoring\n", client.ID)
			continue
		}
		if len(data) == 0 {
			log.Printf("Client %s sent empty message, ignoring\n", client.ID)
			continue
		}

		// message should be in format {"type": "move", "payload": {"direction": "up"}}
		var action game.Action
		err = json.Unmarshal(data, &action)
		if err != nil {
			log.Printf("Error parsing message from client %s: %v\n", client.ID, err)
			continue
		}
		action.ClientID = client.ID   // ensure the action has the correct client ID
		hub.IncomingActions <- action // send the action to the hub for processing
	}
}
