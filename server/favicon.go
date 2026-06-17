package main

import (
	_ "embed"
	"net/http"
)

//go:embed favicon.png
var faviconPNG []byte

// faviconHandler sirve el icono de la app para la pestaña del navegador.
func faviconHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(faviconPNG)
}
