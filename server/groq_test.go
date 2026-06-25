package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGroqClientEnviaLosMensajes verifica que el cliente envía a Groq
// exactamente los mensajes que se le pasan (la personalidad la pone el handler).
func TestGroqClientEnviaLosMensajes(t *testing.T) {
	var recibido groqRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&recibido)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"¡Hola! 🙂"}}]}`))
	}))
	defer server.Close()

	client := NewGroqClient("clave-de-prueba")
	client.URL = server.URL

	entrada := []Message{
		{Role: "system", Content: "eres amable"},
		{Role: "user", Content: "hola"},
	}
	reply, err := client.Complete(context.Background(), entrada, "inteligente")
	if err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}
	if reply != "¡Hola! 🙂" {
		t.Fatalf("respuesta inesperada: %q", reply)
	}
	if len(recibido.Messages) != 2 {
		t.Fatalf("esperaba 2 mensajes, recibí %d", len(recibido.Messages))
	}
	if recibido.Messages[0].Role != "system" || recibido.Messages[1].Content != "hola" {
		t.Fatalf("no se enviaron los mensajes tal cual: %+v", recibido.Messages)
	}
}

// TestPromptDeModo verifica que cada personalidad da un texto distinto y que
// un modo desconocido cae en "amigable".
func TestPromptDeModo(t *testing.T) {
	if promptDeModo("profesor") == promptDeModo("gracioso") {
		t.Fatal("modos distintos deberían dar personalidades distintas")
	}
	if promptDeModo("loquesea") != promptDeModo("amigable") {
		t.Fatal("un modo desconocido debería usar 'amigable'")
	}
}

// TestGroqClientStreamParsea verifica que StreamComplete junta bien los trozos
// que llegan en formato SSE desde Groq.
func TestGroqClientStreamParsea(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hola\"}}]}\n\n")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\" mundo\"}}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client := NewGroqClient("clave-de-prueba")
	client.URL = server.URL

	var juntado string
	err := client.StreamComplete(context.Background(), []Message{{Role: "user", Content: "hola"}}, "inteligente", func(c string) {
		juntado += c
	})
	if err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}
	if juntado != "Hola mundo" {
		t.Fatalf("esperaba 'Hola mundo', obtuve %q", juntado)
	}
}
