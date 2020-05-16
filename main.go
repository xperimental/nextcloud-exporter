package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Version contains the version as set during the build.
	Version = ""

	// GitCommit contains the git commit hash set during the build.
	GitCommit = ""
)

func main() {
	log.Printf("nextcloud-exporter %s", Version)

	config, err := parseConfig()
	if err != nil {
		log.Fatalf("Error in configuration: %s", err)
	}

	log.Printf("Nextcloud server: %s User: %s", config.InfoURL.Hostname(), config.Username)
	collector := newCollector(config.InfoURL, config.Username, config.Password, config.Timeout)
	if err := prometheus.Register(collector); err != nil {
		log.Fatalf("Failed to register collector: %s", err)
	}

	infoMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: metricPrefix + "exporter_info",
		Help: "Information about the nextcloud-exporter.",
		ConstLabels: prometheus.Labels{
			"version": Version,
			"commit":  GitCommit,
		},
	})
	infoMetric.Set(1)
	if err := prometheus.Register(infoMetric); err != nil {
		log.Fatalf("Failed to register info metric: %s", err)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/", http.RedirectHandler("/metrics", http.StatusFound))

	log.Printf("Listen on %s...", config.ListenAddr)
	log.Fatal(http.ListenAndServe(config.ListenAddr, nil))
}
