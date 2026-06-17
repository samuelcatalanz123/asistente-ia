package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// StreamingAIClient es un cliente de IA capaz de entregar la respuesta a trozos.
type StreamingAIClient interface {
	StreamComplete(messages []Message, onChunk func(string)) error
}

// sse escribe un evento Server-Sent Events y lo envía inmediatamente.
func sse(w http.ResponseWriter, flusher http.Flusher, payload any) {
	data, _ := json.Marshal(payload)
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

// NewStreamChatHandler devuelve un handler que transmite la respuesta de la IA
// al navegador trozo a trozo usando Server-Sent Events.
func NewStreamChatHandler(ai StreamingAIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "usa POST"})
			return
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "streaming no disponible"})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "json inválido"})
			return
		}
		if len(req.Messages) == 0 {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "faltan mensajes"})
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		err := ai.StreamComplete(req.Messages, func(chunk string) {
			sse(w, flusher, map[string]string{"t": chunk})
		})
		if err != nil {
			sse(w, flusher, map[string]string{"error": "la IA no respondió, inténtalo de nuevo"})
		}
		// Señal de fin para que el navegador sepa que terminó.
		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	}
}
