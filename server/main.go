package main

import (
	"log"
	"net/http"
	"os"
)

// withCORS allows the Flutter app (any origin) to call the API.
func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func main() {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		log.Fatal("falta la variable GROQ_API_KEY")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var ai AIClient = NewGroqClient(apiKey)

	http.HandleFunc("/health", withCORS(healthHandler))
	http.HandleFunc("/chat", withCORS(NewChatHandler(ai)))

	log.Printf("servidor escuchando en :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
