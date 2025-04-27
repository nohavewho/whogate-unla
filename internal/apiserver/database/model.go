package database

import "time"

// Session represents a chat session
type Session struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Title     string    `json:"title"`
}

// Message represents a chat message
type Message struct {
	ID         string    `json:"id"`
	SessionID  string    `json:"session_id"`
	Content    string    `json:"content"`
	Sender     string    `json:"sender"`
	Timestamp  time.Time `json:"timestamp"`
	ToolCalls  string    `json:"toolCalls,omitempty"`
	ToolResult string    `json:"toolResult,omitempty"`
}
