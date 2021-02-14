package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/xperimental/nextcloud-exporter/internal/config"
	"github.com/xperimental/nextcloud-exporter/internal/login"
	"github.com/xperimental/nextcloud-exporter/serverinfo"
)

var (
	// Version contains the version as set during the build.
	Version = ""

	// GitCommit contains the git commit hash set during the build.
	GitCommit = ""

	log = &logrus.Logger{
		Out: os.Stderr,
		Formatter: &logrus.TextFormatter{
			DisableTimestamp: true,
		},
		Hooks:        make(logrus.LevelHooks),
		Level:        logrus.InfoLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}
)

func main() {
	cfg, err := config.Get()
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err)
	}

	if cfg.RunMode == config.RunModeHelp {
		return
	}

	log.Infof("nextcloud-exporter %s", Version)
	userAgent := fmt.Sprintf("nextcloud-exporter/%s", Version)

	if cfg.RunMode == config.RunModeLogin {
		if cfg.ServerURL == "" {
			log.Fatalf("Need to specify --server for login.")
		}
		loginClient := login.Init(log, userAgent, cfg.ServerURL)

		log.Infof("Starting interactive login on: %s", cfg.ServerURL)
		if err := loginClient.StartInteractive(); err != nil {
			log.Fatalf("Error during login: %s", err)
		}
		return
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %s", err)
	}

	log.Infof("Nextcloud server: %s User: %s", cfg.ServerURL, cfg.Username)

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

	log.Infof("Listen on %s...", cfg.ListenAddr)
	log.Fatal(http.ListenAndServe(cfg.ListenAddr, nil))
}
