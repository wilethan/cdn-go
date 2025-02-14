package server

import (
	"github.com/wilethan/cdn-go/internal/proxy"
	"github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
	"fmt"
	"log"
	"net/http"
	"time"
)

var metricsRegistered = false

var (
    requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cdn_requests_total",
			Help: "Nombre total de requêtes reçues",
		},
		[]string{"method", "path"},
	)

    responseDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cdn_response_duration_seconds",
			Help:    "Durée des réponses HTTP",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func init() {
    prometheus.MustRegister(requestsTotal)
    prometheus.MustRegister(responseDuration)
}

// Middleware pour journaliser les requêtes HTTP
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		// log.Printf("%s %s %s [%s]", r.RemoteAddr, r.Method, r.URL.Path, time.Since(start))
		duration := time.Since(start).Seconds()

		// Mise à jour des métriques Prometheus
		requestsTotal.WithLabelValues(r.Method, r.URL.Path).Inc()
		responseDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)

		log.Printf("%s %s %s [%f s]", r.RemoteAddr, r.Method, r.URL.Path, duration)
	})
}

func prometheusMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        timer := prometheus.NewTimer(responseDuration.WithLabelValues(r.Method, r.URL.Path))
        defer timer.ObserveDuration()

        requestsTotal.WithLabelValues(r.Method, r.URL.Path).Inc()
        next.ServeHTTP(w, r)
    })
}

// Nouvelle fonction pour gérer /get
func getHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("Requête reçue sur /get")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Réponse du serveur pour /get"))
}

func Start() {
    mux := http.NewServeMux()

	// prometheus.MustRegister(requestsTotal)
	// prometheus.MustRegister(responseDuration)

    // Route de santé
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })

	mux.Handle("/metrics", promhttp.Handler())

    // 🔹 Ajout de la route /get
    mux.HandleFunc("/get", getHandler)

    // Création du proxy avec cache
    backendURL := "https://httpbin.org"
    proxyHandler, err := proxy.NewProxy(backendURL, 100)
    if err != nil {
        log.Fatalf("Erreur lors de la création du proxy : %v", err)
    }

    // Route pour relayer les requêtes via le proxy
    mux.Handle("/", proxyHandler)

    // Ajout du middleware de logs
    loggedMux := loggingMiddleware(prometheusMiddleware(mux))

    port := 8080
    fmt.Printf("Serveur proxy avec cache en cours d'exécution sur : http://localhost:%d\n", port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), loggedMux))
}

