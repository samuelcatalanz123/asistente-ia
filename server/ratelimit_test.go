package main

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Bajo muchas peticiones a la vez (misma IP, misma ventana), el límite debe
// respetarse EXACTAMENTE. Corre con -race para detectar condiciones de carrera.
func TestRateLimiterConcurrencia(t *testing.T) {
	const limite = 100
	rl := newRateLimiter(limite, time.Minute)
	ahora := time.Now()

	const goroutines = 50
	const porGoroutine = 10 // 500 intentos en total, muy por encima del límite

	var wg sync.WaitGroup
	var permitidas int64
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < porGoroutine; j++ {
				if rl.permitido("9.9.9.9", ahora) {
					atomic.AddInt64(&permitidas, 1)
				}
			}
		}()
	}
	wg.Wait()

	if permitidas != limite {
		t.Fatalf("con límite %d y 500 intentos concurrentes, esperaba exactamente %d permitidas, obtuve %d",
			limite, limite, permitidas)
	}
}

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
