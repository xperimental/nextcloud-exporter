package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	config, err := parseConfig()
	if err != nil {
		log.Fatalf("Error in configuration: %s", err)
	}

	log.Printf("Nextcloud server: %s User: %s", config.InfoURL.Hostname(), config.Username)
	collector := newCollector(config.InfoURL, config.Username, config.Password, config.Timeout)
	if err := prometheus.Register(collector); err != nil {
		log.Fatalf("Failed to register collector: %s", err)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/", http.RedirectHandler("/metrics", http.StatusFound))

	log.Printf("Listen on %s...", config.ListenAddr)
	log.Fatal(http.ListenAndServe(config.ListenAddr, nil))
}
