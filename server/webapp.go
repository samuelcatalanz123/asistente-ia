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

//go:embed sw.js
var serviceWorker []byte

func manifestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/manifest+json")
	_, _ = w.Write(manifestJSON)
}

// swHandler sirve el service worker que permite instalar la app y usarla sin internet.
func swHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache") // que el navegador note los cambios del SW
	_, _ = w.Write(serviceWorker)
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
