package main

import (
	_ "embed"
	"net/http"
)

// Archivos para que la web se pueda "instalar" en el teléfono (PWA).
//
//go:embed manifest.webmanifest
var manifestJSON []byte

//go:embed icon-192.png
var icon192 []byte

//go:embed icon-512.png
var icon512 []byte

func manifestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/manifest+json")
	_, _ = w.Write(manifestJSON)
}

func icon192Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(icon192)
}

func icon512Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(icon512)
}
