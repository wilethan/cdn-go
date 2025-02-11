package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// NewProxy crée un reverse proxy vers une URL cible
func NewProxy(target string) (*httputil.ReverseProxy, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Intercepter les requêtes avant de les envoyer au backend
	proxy.ModifyResponse = func(resp *http.Response) error {
		log.Printf("Proxy: %s %s -> %d", resp.Request.Method, resp.Request.URL, resp.StatusCode)
		return nil
	}

	return proxy, nil
}
