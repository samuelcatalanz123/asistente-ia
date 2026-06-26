package main

import (
	"context"
	"errors"
	"testing"
)

func TestEsErrorTemporal(t *testing.T) {
	casos := []struct {
		err      error
		temporal bool
	}{
		{nil, false},
		{errors.New("groq 503: server error"), true},
		{errors.New("groq 429: rate limited"), true},
		{errors.New("dial tcp: connection refused"), true},
		{errors.New("groq 400: bad request"), false},
		{errors.New("groq 401: invalid api key"), false},
	}
	for _, c := range casos {
		if got := esErrorTemporal(c.err); got != c.temporal {
			t.Errorf("esErrorTemporal(%v) = %v, quería %v", c.err, got, c.temporal)
		}
	}
}

// Ante errores temporales, debe reintentar hasta tener éxito.
func TestReintentarTextoReintentaYTriunfa(t *testing.T) {
	llamadas := 0
	res, err := reintentarTexto(context.Background(), 3, func() (string, error) {
		llamadas++
		if llamadas < 3 {
			return "", errors.New("groq 503: temporal")
		}
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("esperaba éxito, obtuve error: %v", err)
	}
	if res != "ok" || llamadas != 3 {
		t.Fatalf("res=%q llamadas=%d (quería ok/3)", res, llamadas)
	}
}

// Ante un error NO temporal, debe rendirse a la primera (no reintentar).
func TestReintentarTextoNoReintentaErrorPermanente(t *testing.T) {
	llamadas := 0
	_, err := reintentarTexto(context.Background(), 3, func() (string, error) {
		llamadas++
		return "", errors.New("groq 400: bad request")
	})
	if err == nil {
		t.Fatal("esperaba error permanente")
	}
	if llamadas != 1 {
		t.Fatalf("debía intentar 1 sola vez, intentó %d", llamadas)
	}
}
