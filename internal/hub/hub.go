package hub

import (
	"encoding/json"
	"log"
	"time"

	"github.com/mocci/go-websocket/internal/models"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*models.Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *models.Client

	// Unregister requests from clients
	unregister chan *models.Client
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *models.Client),
		unregister: make(chan *models.Client),
		clients:    make(map[*models.Client]bool),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client registered: %s (%s)", client.Username, client.ID)

			// Broadcast join notification to ALL clients (including the one who just joined)
			joinMsg := models.Message{
				Username:  client.Username,
				Content:   client.Username + " joined the chat",
				Timestamp: time.Now(),
				Type:      "join",
			}

			// Small delay to ensure WritePump is ready
			go func(msg models.Message) {
				time.Sleep(100 * time.Millisecond)
				h.broadcastMessage(msg)
			}(joinMsg)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				log.Printf("Client unregistered: %s (%s)", client.Username, client.ID)

				// Broadcast leave notification
				leaveMsg := models.Message{
					Username:  client.Username,
					Content:   client.Username + " left the chat",
					Timestamp: time.Now(),
					Type:      "leave",
				}
				h.broadcastMessage(leaveMsg)
			}

		case message := <-h.broadcast:
			// Parse the incoming message
			var msg models.Message
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Error parsing message: %v", err)
				continue
			}

			// Set timestamp if not provided
			if msg.Timestamp.IsZero() {
				msg.Timestamp = time.Now()
			}

			// Set type if not provided
			if msg.Type == "" {
				msg.Type = "message"
			}

			// Marshal back to JSON
			msgJSON, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}

			// Send to all connected clients
			for client := range h.clients {
				select {
				case client.Send <- msgJSON:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// Broadcast sends a message to all clients
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

// BroadcastFromClient sends a message from a specific client, auto-injecting username
func (h *Hub) BroadcastFromClient(message []byte, client *models.Client) {
	// Parse the incoming message
	var msg models.Message
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Error parsing message from client: %v", err)
		return
	}

	// SECURITY: Override username from client object (prevent spoofing)
	msg.Username = client.Username

	// Set timestamp if not provided
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	// Set type if not provided
	if msg.Type == "" {
		msg.Type = "message"
	}

	// Marshal back to JSON
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	log.Printf("Broadcasting from %s: %s", client.Username, msg.Content)

	// Broadcast to all clients
	h.broadcast <- msgJSON
}

// Register adds a client to the hub
func (h *Hub) Register(client *models.Client) {
	h.register <- client
}

// Unregister removes a client from the hub
func (h *Hub) Unregister(client *models.Client) {
	h.unregister <- client
}

// GetClients returns the number of connected clients
func (h *Hub) GetClients() int {
	return len(h.clients)
}

// GetUnregisterChan returns the unregister channel
func (h *Hub) GetUnregisterChan() chan<- *models.Client {
	return h.unregister
}

// broadcastMessage is a helper to broadcast a Message struct
func (h *Hub) broadcastMessage(msg models.Message) {
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}
	h.Broadcast(msgJSON)
}
