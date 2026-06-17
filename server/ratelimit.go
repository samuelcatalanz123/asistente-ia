package main

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// visitante guarda cuántas peticiones lleva una IP en la ventana actual.
type visitante struct {
	contador      int
	ventanaInicio time.Time
}

// rateLimiter limita las peticiones por IP usando una ventana de tiempo fija.
type rateLimiter struct {
	mu       sync.Mutex
	visitas  map[string]*visitante
	limite   int
	ventana  time.Duration
}

func newRateLimiter(limite int, ventana time.Duration) *rateLimiter {
	return &rateLimiter{
		visitas: make(map[string]*visitante),
		limite:  limite,
		ventana: ventana,
	}
}

// clientIP saca la IP real del cliente (Render pone la original en X-Forwarded-For).
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// permitido decide si una IP puede hacer otra petición en este momento.
func (rl *rateLimiter) permitido(ip string, ahora time.Time) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, ok := rl.visitas[ip]
	if !ok || ahora.Sub(v.ventanaInicio) > rl.ventana {
		rl.visitas[ip] = &visitante{contador: 1, ventanaInicio: ahora}
		return true
	}
	if v.contador >= rl.limite {
		return false
	}
	v.contador++
	return true
}

// middleware envuelve un handler y rechaza con 429 si se supera el límite.
func (rl *rateLimiter) middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !rl.permitido(clientIP(r), time.Now()) {
			writeJSON(w, http.StatusTooManyRequests,
				ErrorResponse{Error: "demasiadas peticiones, espera un momento"})
			return
		}
		next(w, r)
	}
}
