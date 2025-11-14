package models

import "time"

// Message represents a chat message
type Message struct {
	MessageID string    `json:"messageId,omitempty"` // Client-generated ID for tracking
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // "message", "join", "leave"
}
