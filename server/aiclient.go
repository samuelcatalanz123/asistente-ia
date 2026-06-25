package main

import "context"

// AIClient talks to an AI provider. The HTTP handler depends on this
// interface so tests can substitute a fake without real network calls.
//
// Recibe un context.Context para poder cancelar la petición a la IA si el
// cliente se va (cierra la pestaña, pulsa "parar", etc.).
type AIClient interface {
	Complete(ctx context.Context, messages []Message, modelo string) (string, error)
}
