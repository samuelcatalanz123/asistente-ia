package main

import (
	"log"
	"net/http"
)

// withRecovery atrapa cualquier "pánico" (un fallo grave e inesperado) dentro de
// un handler, lo registra y responde con un 500 limpio. Así un error puntual no
// tumba la conexión ni el servidor: el resto de la app sigue funcionando.
func withRecovery(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("pánico recuperado en %s %s: %v", r.Method, r.URL.Path, err)
				writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "error interno del servidor"})
			}
		}()
		next(w, r)
	}
}
