package models

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 8192
)

// Client represents a websocket client connection
type Client struct {
	// The websocket connection
	Conn *websocket.Conn

	// Buffered channel of outbound messages
	Send chan []byte

	// Client unique ID
	ID string

	// Username for display
	Username string
}

// NewClient creates a new client instance
func NewClient(conn *websocket.Conn, id string, username string) *Client {
	return &Client{
		Conn:     conn,
		Send:     make(chan []byte, 256),
		ID:       id,
		Username: username,
	}
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump(hub interface{ BroadcastFromClient([]byte, *Client) }, unregister chan<- *Client) {
	defer func() {
		log.Printf("ReadPump: Client %s (%s) disconnecting", c.Username, c.ID)
		unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	log.Printf("ReadPump: Client %s (%s) started reading", c.Username, c.ID)

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ReadPump: Unexpected close error for %s: %v", c.Username, err)
			} else {
				log.Printf("ReadPump: Client %s read error: %v", c.Username, err)
			}
			break
		}

		log.Printf("ReadPump: Client %s received message: %s", c.Username, string(message))
		// Broadcast message with client context (server will inject username)
		hub.BroadcastFromClient(message, c)
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
