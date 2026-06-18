package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeStreamer emite una lista fija de trozos.
type fakeStreamer struct {
	chunks []string
	got    []Message
}

func (f *fakeStreamer) StreamComplete(messages []Message, onChunk func(string)) error {
	f.got = messages
	for _, c := range f.chunks {
		onChunk(c)
	}
	return nil
}

func TestStreamHandlerEnviaTrozos(t *testing.T) {
	fake := &fakeStreamer{chunks: []string{"Hola", " mundo", "!"}}
	handler := NewStreamChatHandler(fake)

	body := `{"messages":[{"role":"user","content":"saluda"}]}`
	req := httptest.NewRequest(http.MethodPost, "/chat/stream", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d", rec.Code)
	}
	salida := rec.Body.String()
	// Cada trozo debe ir como un evento SSE con su texto.
	for _, esperado := range []string{`"t":"Hola"`, `"t":" mundo"`, `"t":"!"`, "[DONE]"} {
		if !strings.Contains(salida, esperado) {
			t.Fatalf("la salida no contiene %q. Salida: %s", esperado, salida)
		}
	}
	// El handler antepone la personalidad: llega [system, user].
	if len(fake.got) != 2 || fake.got[0].Role != "system" || fake.got[1].Content != "saluda" {
		t.Fatalf("no se pasaron bien los mensajes (system + user): %+v", fake.got)
	}
}

func TestStreamHandlerRechazaVacio(t *testing.T) {
	handler := NewStreamChatHandler(&fakeStreamer{})
	req := httptest.NewRequest(http.MethodPost, "/chat/stream", strings.NewReader(`{"messages":[]}`))
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperaba 400 con mensajes vacíos, obtuve %d", rec.Code)
	}
}
