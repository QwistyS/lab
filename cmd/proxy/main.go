package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
	// Read target from env
	target := os.Getenv("TARGET_URL")
	if target == "" {
		log.Fatal("TARGET_URL env var is required, e.g. TARGET_URL=http://example.com")
	}

	u, err := url.Parse(target)
	if err != nil {
		log.Fatalf("invalid TARGET_URL %q: %v", target, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(u)

	// Optional: you can tweak the Director if you want to change headers etc.
	origDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		origDirector(r)
		// If upstream cares about Host:
		// r.Host = u.Host
	}

	// Health endpoint for k8s later
	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	// Everything else gets proxied
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s%s", r.Method, r.Host, r.URL.Path)
		proxy.ServeHTTP(w, r)
	})

	addr := ":8080"
	log.Printf("starting proxy on %s -> %s", addr, u.String())
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
