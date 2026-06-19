package main

import (
	"encoding/json"
	"log"
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
		// Limita el cuerpo a 1 MB para que nadie pueda saturar el servidor.
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
		// Anteponemos la personalidad elegida (modo) como mensaje "system".
		mensajes := append([]Message{{Role: "system", Content: promptDeModo(req.Modo)}}, req.Messages...)
		reply, err := ai.Complete(mensajes, req.Modelo)
		if err != nil && req.Modelo != "rapido" {
			// Si el modelo grande falló (p. ej. agotó su cuota), probamos el rápido.
			log.Printf("groq falló con %q, probando 'rapido': %v", req.Modelo, err)
			reply, err = ai.Complete(mensajes, "rapido")
		}
		if err != nil {
			log.Printf("error de groq: %v", err)
			writeJSON(w, http.StatusBadGateway, ErrorResponse{
				Error: "El asistente está ocupado ahora mismo 😅. Espera un momento e inténtalo de nuevo."})
			return
		}
		writeJSON(w, http.StatusOK, ChatResponse{Reply: reply})
	}
}
