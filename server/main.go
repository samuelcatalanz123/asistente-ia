package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	groq := NewGroqClient(apiKey)

	// Máximo 30 peticiones por minuto por IP en las rutas del chat.
	limitador := newRateLimiter(30, time.Minute)

	http.HandleFunc("/", withCORS(homeHandler))
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/favicon.png", faviconHandler)
	http.HandleFunc("/manifest.webmanifest", manifestHandler)
	http.HandleFunc("/icon-192.png", icon192Handler)
	http.HandleFunc("/icon-512.png", icon512Handler)
	http.HandleFunc("/sw.js", swHandler)
	http.HandleFunc("/health", withLogging(withRecovery(withCORS(healthHandler))))
	http.HandleFunc("/chat", withLogging(withRecovery(withCORS(limitador.middleware(NewChatHandler(groq))))))
	http.HandleFunc("/chat/stream", withLogging(withRecovery(withCORS(limitador.middleware(NewStreamChatHandler(groq))))))

	srv := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 35 * time.Second, // un poco más que el timeout de Groq (30s)
	}

	// El servidor escucha en segundo plano para poder atender la señal de apagado.
	go func() {
		log.Printf("servidor escuchando en :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error del servidor: %v", err)
		}
	}()

	// Apagado elegante: cuando el sistema pide cerrar (p. ej. Render en cada
	// despliegue manda SIGTERM), dejamos terminar las peticiones en curso.
	parar := make(chan os.Signal, 1)
	signal.Notify(parar, os.Interrupt, syscall.SIGTERM)
	<-parar
	log.Println("apagando el servidor (terminando peticiones en curso)...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("apagado forzado: %v", err)
	}
	log.Println("servidor detenido limpiamente ✅")
}
