package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	healthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", body["status"])
	}
}

// /health debe incluir la versión y el tiempo activo (útil para monitoreo).
func TestHealthHandlerIncludesVersionAndUptime(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	healthHandler(rec, req)

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body["version"] == "" {
		t.Fatal("health no incluye la versión")
	}
	if body["uptime"] == "" {
		t.Fatal("health no incluye el tiempo activo (uptime)")
	}
}

// fakeAI is a test double for AIClient.
type fakeAI struct {
	reply  string
	err    error
	got    []Message
	gotCtx context.Context
}

func (f *fakeAI) Complete(ctx context.Context, messages []Message, modelo string) (string, error) {
	f.gotCtx = ctx
	f.got = messages
	return f.reply, f.err
}

// El handler debe propagar el context de la petición HTTP hasta la llamada a la
// IA (así, si el cliente se va, se puede cancelar la petición a Groq).
func TestChatHandlerPropagaContext(t *testing.T) {
	fake := &fakeAI{reply: "ok"}
	handler := NewChatHandler(fake)

	type clave struct{}
	ctx := context.WithValue(context.Background(), clave{}, "sí")
	req := httptest.NewRequest(http.MethodPost, "/chat",
		strings.NewReader(`{"messages":[{"role":"user","content":"hola"}]}`)).WithContext(ctx)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if fake.gotCtx == nil {
		t.Fatal("el handler no propagó ningún context a Complete")
	}
	if fake.gotCtx.Value(clave{}) != "sí" {
		t.Fatal("el context propagado no es el de la petición HTTP")
	}
}

func TestChatHandlerReturnsReply(t *testing.T) {
	fake := &fakeAI{reply: "¡Hola! ¿En qué te ayudo?"}
	handler := NewChatHandler(fake)

	body := `{"messages":[{"role":"user","content":"hola"}]}`
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var resp ChatResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if resp.Reply != "¡Hola! ¿En qué te ayudo?" {
		t.Fatalf("unexpected reply: %q", resp.Reply)
	}
	// El handler antepone la personalidad: llega [system, user].
	if len(fake.got) != 2 {
		t.Fatalf("esperaba 2 mensajes (system + user), got %+v", fake.got)
	}
	if fake.got[0].Role != "system" {
		t.Fatalf("el primer mensaje debe ser la personalidad (system): %+v", fake.got[0])
	}
	if fake.got[1].Content != "hola" {
		t.Fatalf("no se conservó el mensaje del usuario: %+v", fake.got)
	}
}

func TestChatHandlerRejectsEmptyMessages(t *testing.T) {
	handler := NewChatHandler(&fakeAI{reply: "x"})
	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{"messages":[]}`))
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty messages, got %d", rec.Code)
	}
}

func TestChatHandlerRejectsNonPost(t *testing.T) {
	handler := NewChatHandler(&fakeAI{reply: "x"})
	req := httptest.NewRequest(http.MethodGet, "/chat", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for GET, got %d", rec.Code)
	}
}

func TestChatHandlerReturns502WhenAIFails(t *testing.T) {
	handler := NewChatHandler(&fakeAI{err: errors.New("boom")})
	body := `{"messages":[{"role":"user","content":"hola"}]}`
	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 when AI fails, got %d", rec.Code)
	}
}
