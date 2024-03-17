package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/xperimental/nextcloud-exporter/internal/client"
	"github.com/xperimental/nextcloud-exporter/internal/config"
	"github.com/xperimental/nextcloud-exporter/internal/login"
	"github.com/xperimental/nextcloud-exporter/internal/metrics"
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

	if cfg.RunMode == config.RunModeVersion {
		fmt.Println(Version)
		return
	}

	log.Infof("nextcloud-exporter %s", Version)
	userAgent := fmt.Sprintf("nextcloud-exporter/%s", Version)

	if cfg.RunMode == config.RunModeLogin {
		if cfg.ServerURL == "" {
			log.Fatalf("Need to specify --server for login.")
		}
		loginClient := login.Init(log, userAgent, cfg.ServerURL, cfg.TLSSkipVerify)

		log.Infof("Starting interactive login on: %s", cfg.ServerURL)
		if err := loginClient.StartInteractive(); err != nil {
			log.Fatalf("Error during login: %s", err)
		}
		return
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %s", err)
	}

	if cfg.AuthToken == "" {
		log.Infof("Nextcloud server: %s User: %s", cfg.ServerURL, cfg.Username)
	} else {
		log.Infof("Nextcloud server: %s Authentication using token.", cfg.ServerURL)
	}

	infoURL := serverinfo.InfoURL(cfg.ServerURL, !cfg.Info.Apps)

	if cfg.TLSSkipVerify {
		log.Warn("HTTPS certificate verification is disabled.")
	}

	infoClient := client.New(infoURL, cfg.Username, cfg.Password, cfg.AuthToken, cfg.Timeout, userAgent, cfg.TLSSkipVerify)
	if err := metrics.RegisterCollector(log, infoClient, cfg.Info.Apps); err != nil {
		log.Fatalf("Failed to register collector: %s", err)
	}

	if err := metrics.RegisterInfoMetric(Version, GitCommit); err != nil {
		log.Fatalf("Failed to register info metric: %s", err)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/", http.RedirectHandler("/metrics", http.StatusFound))

	log.Infof("Listen on %s...", cfg.ListenAddr)
	log.Fatal(http.ListenAndServe(cfg.ListenAddr, nil))
}
