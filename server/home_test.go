package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHomeHandlerServesPage(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	homeHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Asistente IA") {
		t.Fatalf("home page does not contain the title")
	}
}

func TestHomeHandler404OnUnknownPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/cualquier-cosa", nil)
	rec := httptest.NewRecorder()

	homeHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown path, got %d", rec.Code)
	}
}
