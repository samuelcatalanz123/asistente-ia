package main

// AIClient talks to an AI provider. The HTTP handler depends on this
// interface so tests can substitute a fake without real network calls.
type AIClient interface {
	Complete(messages []Message) (string, error)
}
