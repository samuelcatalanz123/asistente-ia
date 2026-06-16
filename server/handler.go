package main

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// NewChatHandler returns an http handler that forwards the conversation
// to the given AIClient and returns the reply as JSON.
func NewChatHandler(ai AIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "usa POST"})
			return
		}
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "json inválido"})
			return
		}
		if len(req.Messages) == 0 {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "faltan mensajes"})
			return
		}
		reply, err := ai.Complete(req.Messages)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, ErrorResponse{Error: "la IA no respondió, inténtalo de nuevo"})
			return
		}
		writeJSON(w, http.StatusOK, ChatResponse{Reply: reply})
	}
}
