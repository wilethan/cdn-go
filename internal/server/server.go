package server

import (
	"github.com/wilethan/cdn-go/internal/proxy"
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

// Start lance le serveur HTTP avec proxy et middleware
func Start() {
	mux := http.NewServeMux()

	// Route de santé
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Création du proxy HTTP vers un backend (exemple : httpbin.org)
	backendURL := "https://httpbin.org"
	proxyHandler, err := proxy.NewProxy(backendURL)
	if err != nil {
		log.Fatalf("Erreur lors de la création du proxy : %v", err)
	}

	// Route pour relayer les requêtes vers le backend
	mux.Handle("/", proxyHandler)

	// Ajout du middleware de logs
	loggedMux := loggingMiddleware(mux)

	port := 8080
	fmt.Printf("Serveur proxy en cours d'exécution sur : http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), loggedMux))
}
