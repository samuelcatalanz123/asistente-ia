package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// appVersion identifica la versión desplegada del servidor.
const appVersion = "1.0.3"

// arranque guarda cuándo empezó el servidor, para calcular el tiempo activo.
var arranque = time.Now()

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// healthHandler informa el estado del servicio (lo usan las herramientas de monitoreo).
func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": appVersion,
		"uptime":  time.Since(arranque).Round(time.Second).String(),
	})
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
		// Reintentamos hasta 3 veces si el error es temporal (429/5xx/red), con espera creciente.
		reply, err := reintentarTexto(r.Context(), 3, func() (string, error) {
			return ai.Complete(r.Context(), mensajes, req.Modelo)
		})
		if err != nil && req.Modelo != "rapido" && r.Context().Err() == nil {
			// Si el modelo grande falló (p. ej. agotó su cuota), probamos el rápido.
			// (Salvo que el cliente ya se haya ido: ahí no reintentamos.)
			log.Printf("groq falló con %q, probando 'rapido': %v", req.Modelo, err)
			reply, err = reintentarTexto(r.Context(), 2, func() (string, error) {
				return ai.Complete(r.Context(), mensajes, "rapido")
			})
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
