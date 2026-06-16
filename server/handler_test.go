package main

import (
	"bytes"
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

// fakeAI is a test double for AIClient.
type fakeAI struct {
	reply string
	err   error
	got   []Message
}

func (f *fakeAI) Complete(messages []Message) (string, error) {
	f.got = messages
	return f.reply, f.err
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
	if len(fake.got) != 1 || fake.got[0].Content != "hola" {
		t.Fatalf("handler did not pass messages through: %+v", fake.got)
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
