package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xperimental/nextcloud-exporter/internal/config"
	"github.com/xperimental/nextcloud-exporter/internal/login"
	"github.com/xperimental/nextcloud-exporter/serverinfo"
)

var (
	// Version contains the version as set during the build.
	Version = ""

	// GitCommit contains the git commit hash set during the build.
	GitCommit = ""
)

func main() {
	log.SetFlags(0)
	log.Printf("nextcloud-exporter %s", Version)
	userAgent := fmt.Sprintf("nextcloud-exporter/%s", Version)

	cfg, err := config.Get()
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err)
	}

	if err := cfg.Validate(); err != nil {
		if cfg.LoginMode() {
			log.Printf("Starting interactive login for %q on: %s", cfg.Username, cfg.ServerURL)
			if err := login.StartInteractive(userAgent, cfg.ServerURL, cfg.Username); err != nil {
				log.Fatalf("Error during login: %s", err)
			}
			return
		}

		log.Fatalf("Invalid configuration: %s", err)
	}

	log.Printf("Nextcloud server: %s User: %s", cfg.ServerURL, cfg.Username)

	infoURL := cfg.ServerURL + serverinfo.InfoPath

	collector := newCollector(infoURL, cfg.Username, cfg.Password, cfg.Timeout, userAgent)
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

	log.Printf("Listen on %s...", cfg.ListenAddr)
	log.Fatal(http.ListenAndServe(cfg.ListenAddr, nil))
}
