package main

import (
	"context"
	"strings"
	"time"
)

// esErrorTemporal indica si un error de la IA es pasajero y vale la pena
// reintentar: límite de peticiones (429), errores del servidor (5xx) o de red.
// Un 400/401 (petición mal hecha o clave inválida) NO es temporal.
func esErrorTemporal(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	for _, marca := range []string{"groq 5", "groq 429", "timeout", "connection", "eof", "reset", "no such host"} {
		if strings.Contains(s, marca) {
			return true
		}
	}
	return false
}

// reintentarTexto llama a fn hasta `intentos` veces mientras devuelva un error
// temporal, esperando un poco más cada vez (backoff). Se detiene de inmediato si
// el error no es temporal, si tiene éxito, o si el cliente se va (ctx cancelado).
func reintentarTexto(ctx context.Context, intentos int, fn func() (string, error)) (string, error) {
	var res string
	var err error
	for i := 0; i < intentos; i++ {
		res, err = fn()
		if err == nil || !esErrorTemporal(err) || ctx.Err() != nil {
			return res, err
		}
		select {
		case <-ctx.Done():
			return res, err
		case <-time.After(time.Duration(i+1) * 300 * time.Millisecond): // 300ms, 600ms, ...
		}
	}
	return res, err
}
