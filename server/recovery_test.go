package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Si un handler entra en pánico, el middleware debe atraparlo y responder 500
// (no debe propagar el pánico ni tumbar el proceso).
func TestWithRecoveryAtrapaPanico(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		panic("algo explotó")
	}
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()

	withRecovery(handler)(rec, req) // si propagara el pánico, el test fallaría aquí

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("esperaba 500 tras el pánico, obtuve %d", rec.Code)
	}
}

// Sin pánico, el middleware debe dejar pasar la respuesta normal intacta.
func TestWithRecoveryDejaPasarRespuestaNormal(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()

	withRecovery(handler)(rec, req)

	if rec.Code != http.StatusOK || rec.Body.String() != "ok" {
		t.Fatalf("no dejó pasar la respuesta normal: %d %q", rec.Code, rec.Body.String())
	}
}
