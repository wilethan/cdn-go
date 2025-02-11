package server

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// Middleware pour journaliser les requêtes HTTP
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s [%s]", r.RemoteAddr, r.Method, r.URL.Path, time.Since(start))
	})
}

// Start lance le serveur HTTP avec le middleware de logs
func Start() {
	mux := http.NewServeMux()

	// Route de santé
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Ajout du middleware de logs
	loggedMux := loggingMiddleware(mux)

	port := 8080
	fmt.Printf("Serveur en cours d'exécution sur : http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), loggedMux))
}
