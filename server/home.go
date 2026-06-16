package main

import (
	_ "embed"
	"net/http"
)

//go:embed home.html
var homePage []byte

// homeHandler serves the web chat page at the root URL so the assistant
// can be used directly from a browser.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Solo la raíz "/" muestra la página; cualquier otra ruta es 404.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(homePage)
}
