package game

import (
	"log"
	"time"
)

type Action struct {
	ClientID string
	Data     string
}

type CentralHub struct {
	// channels to manage lifecycle
	Register        chan string
	Unregister      chan string
	IncomingActions chan Action

	// internal state
	Clients map[string]bool
}

func NewHub() *CentralHub {
	return &CentralHub{
		Register:        make(chan string),
		Unregister:      make(chan string),
		IncomingActions: make(chan Action),
		Clients:         make(map[string]bool),
	}
}

func (hub *CentralHub) Run() {
	// ticker that fires at a steady interval (e.g., 60 times per second)
	ticker := time.NewTicker(500 * time.Millisecond) // adjust as needed for game speed
	defer ticker.Stop()
	for {
		select {
		case id := <-hub.Register:
			hub.Clients[id] = true
			log.Printf("Client %s registered. Total clients: %d\n", id, len(hub.Clients))
		case id := <-hub.Unregister:
			delete(hub.Clients, id)
			log.Printf("Client %s unregistered. Total clients: %d\n", id, len(hub.Clients))
		case action := <-hub.IncomingActions:
			log.Printf("Processing action from client %s: %s\n", action.ClientID, action.Data)
			// TODO: update state buffer
		case <-ticker.C:
			log.Println("Game tick - simulating world update...")
			// TODO: look at state buffer and execute game rules
			// e.g. move things, check collisions, and broadcast new state to clients
		}
	}
}
