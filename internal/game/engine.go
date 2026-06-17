package game

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Action struct {
	ClientID string
	Type     string          `json:"type"`    // e.g. "move"
	Payload  json.RawMessage `json:"payload"` // e.g. {"direction": "right"}
}

type MoveAction struct {
	ClientID  string
	Direction Position `json:"direction"`
}

type Client struct {
	ID       string
	Conn     *websocket.Conn
	SendChan chan []byte
}

type CentralHub struct {
	Register        chan Client
	Unregister      chan Client
	IncomingActions chan Action
	Clients         map[string]*Client // map Client ID to Client struct
	State           *GameState
}

func NewHub() *CentralHub {
	return &CentralHub{
		Register:        make(chan Client),
		Unregister:      make(chan Client),
		IncomingActions: make(chan Action),
		Clients:         make(map[string]*Client),
		State:           NewGameState(),
	}
}

var directionMap = map[string]Position{
	"UP":    {X: 0, Y: -1},
	"DOWN":  {X: 0, Y: 1},
	"LEFT":  {X: -1, Y: 0},
	"RIGHT": {X: 1, Y: 0},
}

// Parses a MoveAction from an Action struct, validating the direction
func ParseMoveAction(action Action) (*MoveAction, error) {
	if action.Type != "move" {
		return nil, fmt.Errorf("invalid action type: expected 'move', got '%s'", action.Type)
	}
	var payload struct {
		Direction string `json:"direction"`
	}
	err := json.Unmarshal(action.Payload, &payload)
	if err != nil {
		return nil, fmt.Errorf("error parsing move action payload: %v", err)
	}

	var direction Position
	if d, ok := directionMap[strings.ToUpper(payload.Direction)]; ok {
		direction = d
	} else {
		return nil, fmt.Errorf("invalid move direction: %s", payload.Direction)
	}

	return &MoveAction{
		ClientID:  action.ClientID,
		Direction: direction,
	}, nil
}

// Sends marshalled game state to all clients
func (hub *CentralHub) BroadcastState() {
	stateJSON, err := json.Marshal(hub.State)
	if err != nil {
		log.Printf("Error marshalling game state: %v\n", err)
		return
	}

	log.Printf("Broadcasting game state to %d clients: %s\n", len(hub.Clients), stateJSON)
	for id, client := range hub.Clients {
		select {
		case client.SendChan <- stateJSON:
			// send successful
		default:
			log.Printf("Send channel full for client %s, unregistering client\n", id)
			hub.Unregister <- *client
		}
	}
}

func (hub *CentralHub) Run() {
	// ticker that fires at a steady interval (e.g., 60 times per second)
	ticker := time.NewTicker(500 * time.Millisecond) // adjust as needed for game speed
	defer ticker.Stop()

	for {
		select {
		case newClient := <-hub.Register:
			hub.Clients[newClient.ID] = &newClient
			hub.State.AddSnake(newClient.ID) // add new snake for this client at a default position
			log.Printf("Client %s registered. Total clients: %d\n", newClient.ID, len(hub.Clients))

		case client := <-hub.Unregister:
			delete(hub.Clients, client.ID)
			hub.State.RemoveSnake(client.ID) // remove snake for this client
			log.Printf("Client %s unregistered. Total clients: %d\n", client.ID, len(hub.Clients))

		case action := <-hub.IncomingActions:
			log.Printf("Processing %s action from client %s: %s\n", action.Type, action.ClientID, action.Payload)
			if strings.ToLower(action.Type) == "move" {
				moveAction, err := ParseMoveAction(action)
				if err != nil {
					log.Printf("Error parsing move action from client %s: %v\n", action.ClientID, err)
					continue
				}
				err = hub.State.UpdateSnakeDirection(moveAction.ClientID, moveAction.Direction)
				if err != nil {
					log.Printf("Error updating snake direction for client %s: %v\n", moveAction.ClientID, err)
				} else {
					log.Printf("Updated direction for client %s to %v\n", moveAction.ClientID, moveAction.Direction)
				}
			} else {
				log.Printf("Unknown action type %s from client %s\n", action.Type, action.ClientID)
			}

		case <-ticker.C:
			log.Println("Game tick - simulating world update...")
			hub.State.UpdateGameState()
			hub.BroadcastState()
		}
	}
}
