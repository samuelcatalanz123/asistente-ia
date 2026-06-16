package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGroqClientSendsPersonality verifica que el cliente antepone el mensaje
// "system" con la personalidad antes de enviar la conversación a Groq.
func TestGroqClientSendsPersonality(t *testing.T) {
	var recibido groqRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&recibido)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"¡Hola! 🙂"}}]}`))
	}))
	defer server.Close()

	client := NewGroqClient("clave-de-prueba")
	client.URL = server.URL

	reply, err := client.Complete([]Message{{Role: "user", Content: "hola"}})
	if err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}
	if reply != "¡Hola! 🙂" {
		t.Fatalf("respuesta inesperada: %q", reply)
	}

	// El primer mensaje debe ser la personalidad (rol system).
	if len(recibido.Messages) != 2 {
		t.Fatalf("esperaba 2 mensajes (system + user), recibí %d", len(recibido.Messages))
	}
	if recibido.Messages[0].Role != "system" {
		t.Fatalf("el primer mensaje debe ser 'system', fue %q", recibido.Messages[0].Role)
	}
	if recibido.Messages[0].Content != systemPrompt {
		t.Fatalf("la personalidad enviada no coincide")
	}
	if recibido.Messages[1].Role != "user" || recibido.Messages[1].Content != "hola" {
		t.Fatalf("el mensaje del usuario no se conservó: %+v", recibido.Messages[1])
	}
}
