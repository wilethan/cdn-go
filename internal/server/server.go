package server

import (
	"fmt"
	"log"
	"net/http"
)

// Start lance le serveur HTTP
func Start() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	port := 8080
	fmt.Printf("Serveur en cours d'exécution sur : http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
