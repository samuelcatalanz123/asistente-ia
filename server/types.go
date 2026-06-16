package main

// Message is one turn in the conversation.
type Message struct {
	Role    string `json:"role"` // "user" or "assistant"
	Content string `json:"content"`
}

// ChatRequest is what the Flutter app sends to POST /chat.
type ChatRequest struct {
	Messages []Message `json:"messages"`
}

// ChatResponse is what POST /chat returns.
type ChatResponse struct {
	Reply string `json:"reply"`
}

// ErrorResponse is returned on failures.
type ErrorResponse struct {
	Error string `json:"error"`
}
