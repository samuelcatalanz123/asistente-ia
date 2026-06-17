package main

import (
	"testing"
	"time"
)

func TestRateLimiterLimitaPorIP(t *testing.T) {
	rl := newRateLimiter(2, time.Minute)
	base := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)

	if !rl.permitido("1.1.1.1", base) {
		t.Fatal("la 1ª petición debería permitirse")
	}
	if !rl.permitido("1.1.1.1", base) {
		t.Fatal("la 2ª petición debería permitirse")
	}
	if rl.permitido("1.1.1.1", base) {
		t.Fatal("la 3ª petición debería rechazarse (supera el límite)")
	}
}

func TestRateLimiterSeReiniciaTrasLaVentana(t *testing.T) {
	rl := newRateLimiter(1, time.Minute)
	base := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)

	if !rl.permitido("2.2.2.2", base) {
		t.Fatal("la 1ª petición debería permitirse")
	}
	if rl.permitido("2.2.2.2", base) {
		t.Fatal("la 2ª seguida debería rechazarse")
	}
	// Pasado el minuto, vuelve a permitirse.
	if !rl.permitido("2.2.2.2", base.Add(2*time.Minute)) {
		t.Fatal("tras la ventana debería permitirse de nuevo")
	}
}

func TestRateLimiterIndependientePorIP(t *testing.T) {
	rl := newRateLimiter(1, time.Minute)
	base := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)

	if !rl.permitido("3.3.3.3", base) {
		t.Fatal("IP A debería permitirse")
	}
	if !rl.permitido("4.4.4.4", base) {
		t.Fatal("IP B no debería verse afectada por IP A")
	}
}
