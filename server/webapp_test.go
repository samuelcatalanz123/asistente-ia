package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// El service worker debe servirse como JavaScript y traer su lógica de caché.
func TestServiceWorkerHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/sw.js", nil)
	rec := httptest.NewRecorder()

	swHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "javascript") {
		t.Fatalf("Content-Type debería ser javascript, fue %q", ct)
	}
	if !strings.Contains(rec.Body.String(), "addEventListener") {
		t.Fatal("el service worker no contiene su lógica (addEventListener)")
	}
}

// El manifest hace la web instalable: debe ser JSON de manifiesto y declarar standalone.
func TestManifestHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/manifest.webmanifest", nil)
	rec := httptest.NewRecorder()

	manifestHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "manifest") {
		t.Fatalf("Content-Type debería ser de manifest, fue %q", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Asistente IA") {
		t.Fatal("el manifest no trae el nombre de la app")
	}
	if !strings.Contains(body, "standalone") {
		t.Fatal("el manifest no declara display:standalone (requisito para instalar)")
	}
}

// Los iconos de la PWA deben servirse como PNG y no venir vacíos.
func TestIconHandlers(t *testing.T) {
	casos := []struct {
		nombre  string
		ruta    string
		handler http.HandlerFunc
	}{
		{"icon-192", "/icon-192.png", icon192Handler},
		{"icon-512", "/icon-512.png", icon512Handler},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, c.ruta, nil)
			rec := httptest.NewRecorder()

			c.handler(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("esperaba 200, obtuve %d", rec.Code)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
				t.Fatalf("Content-Type debería ser image/png, fue %q", ct)
			}
			if rec.Body.Len() == 0 {
				t.Fatal("el icono vino vacío")
			}
		})
	}
}
