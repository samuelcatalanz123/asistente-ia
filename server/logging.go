package main

import (
	"log"
	"net/http"
	"time"
)

// statusRecorder envuelve el ResponseWriter para recordar con qué código de
// estado respondió el handler (el ResponseWriter normal no lo expone).
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Flush deja pasar el vaciado al ResponseWriter real. Es imprescindible para
// el streaming (SSE): sin esto, el envoltorio rompería http.Flusher.
func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// withLogging registra cada petición: método, ruta, código de estado y duración.
// Es un "middleware" de observabilidad: deja ver qué hace el servidor en vivo.
func withLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next(rec, r)
		log.Printf("%s %s → %d (%s)", r.Method, r.URL.Path, rec.status, time.Since(inicio).Round(time.Millisecond))
	}
}
