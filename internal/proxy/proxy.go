package proxy

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/hashicorp/golang-lru"
)

// Proxy représente un proxy HTTP avec cache
type Proxy struct {
	backendURL *url.URL
	cache      *lru.Cache
}

// NewProxy crée une nouvelle instance de proxy avec cache
func NewProxy(backend string, cacheSize int) (*Proxy, error) {
	parsedURL, err := url.Parse(backend)
	if err != nil {
		return nil, fmt.Errorf("URL backend invalide: %w", err)
	}

	cache, err := lru.New(cacheSize)
	if err != nil {
		return nil, fmt.Errorf("Erreur lors de la création du cache: %w", err)
	}

	return &Proxy{
		backendURL: parsedURL,
		cache:      cache,
	}, nil
}

// ServeHTTP gère les requêtes HTTP entrantes
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Construction de l'URL de destination
	targetURL := p.backendURL.ResolveReference(r.URL)

	// Vérifier si la réponse est en cache
	if cachedResponse, found := p.cache.Get(targetURL.String()); found {
		log.Printf("[CACHE HIT] %s", targetURL.String())
		w.WriteHeader(http.StatusOK)
		w.Write(cachedResponse.([]byte))
		return
	}

	log.Printf("[PROXY] Requête vers %s", targetURL.String())

	// Création d'une nouvelle requête HTTP
	req, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		http.Error(w, "Erreur interne lors de la création de la requête", http.StatusInternalServerError)
		return
	}
	req.Header = r.Header

	// Client HTTP avec timeout et gestion des erreurs TLS
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // À sécuriser en production
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Erreur lors de la requête au serveur distant", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Lire la réponse du backend
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Erreur lors de la lecture de la réponse", http.StatusInternalServerError)
		return
	}

	// Stocker la réponse en cache si elle est de type 200 OK
	if resp.StatusCode == http.StatusOK {
		p.cache.Add(targetURL.String(), body)
	}

	// Copier les headers et écrire la réponse au client
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}
