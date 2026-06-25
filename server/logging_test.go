package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// El envoltorio debe recordar el código de estado y pasarlo al cliente real.
func TestStatusRecorderCapturaEstado(t *testing.T) {
	rec := httptest.NewRecorder()
	sr := &statusRecorder{ResponseWriter: rec, status: http.StatusOK}

	sr.WriteHeader(http.StatusNotFound)

	if sr.status != http.StatusNotFound {
		t.Fatalf("el recorder no guardó el estado: %d", sr.status)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("el estado no llegó al ResponseWriter real: %d", rec.Code)
	}
}

// withLogging no debe alterar la respuesta del handler (ni el estado ni el cuerpo).
func TestWithLoggingNoAlteraRespuesta(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("hola"))
	}
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()

	withLogging(handler)(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("esperaba 418, obtuve %d", rec.Code)
	}
	if rec.Body.String() != "hola" {
		t.Fatalf("el cuerpo cambió: %q", rec.Body.String())
	}
}

// statusRecorder debe seguir siendo un http.Flusher: si no, rompería el streaming (SSE).
func TestStatusRecorderEsFlusher(t *testing.T) {
	var w http.ResponseWriter = &statusRecorder{ResponseWriter: httptest.NewRecorder(), status: 200}
	if _, ok := w.(http.Flusher); !ok {
		t.Fatal("statusRecorder debería implementar http.Flusher (si no, rompe el streaming)")
	}
}
