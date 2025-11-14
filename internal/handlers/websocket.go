package handlers

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/mocci/go-websocket/internal/hub"
	"github.com/mocci/go-websocket/internal/models"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for development (change this in production!)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocketHandler handles websocket requests from the client
func WebSocketHandler(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get username from query parameter
		username := r.URL.Query().Get("username")
		if username == "" {
			username = "Anonymous"
		}

		// Upgrade HTTP connection to WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Error upgrading connection: %v", err)
			return
		}

		// Create new client
		clientID := uuid.New().String()
		client := models.NewClient(conn, clientID, username)

		// Register client with hub
		h.Register(client)

		// Start goroutines for reading and writing
		go client.WritePump()
		go client.ReadPump(h, h.GetUnregisterChan())

		log.Printf("New WebSocket connection: %s (%s)", username, clientID)
	}
}

