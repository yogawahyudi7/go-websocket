package main

import (
	"log"
	"net/http"

	"github.com/mocci/go-websocket/internal/handlers"
	"github.com/mocci/go-websocket/internal/hub"
)

func main() {
	// Create new hub
	h := hub.NewHub()

	// Start hub in a goroutine
	go h.Run()

	// Setup WebSocket route
	http.HandleFunc("/ws", handlers.WebSocketHandler(h))

	// Start server
	addr := ":8080"
	log.Printf("WebSocket server starting on %s", addr)
	log.Printf("Connect to: ws://localhost%s/ws?username=YourName", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
