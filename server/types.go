package main

// Message is one turn in the conversation.
type Message struct {
	Role    string `json:"role"` // "user" or "assistant"
	Content string `json:"content"`
	// Imagen es una foto en base64 (data URL) que el usuario adjunta para que
	// la IA la "vea". Vacío en los mensajes normales.
	Imagen string `json:"imagen,omitempty"`
}

// ChatRequest is what the Flutter app sends to POST /chat.
type ChatRequest struct {
	Messages []Message `json:"messages"`
	Modo     string    `json:"modo,omitempty"`   // personalidad: amigable, profesor, programador, gracioso
	Modelo   string    `json:"modelo,omitempty"` // cerebro: rapido o inteligente
}

// ChatResponse is what POST /chat returns.
type ChatResponse struct {
	Reply string `json:"reply"`
}

// ErrorResponse is returned on failures.
type ErrorResponse struct {
	Error string `json:"error"`
}
