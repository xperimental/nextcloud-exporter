package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
)

type config struct {
	ListenAddr string
	Timeout    time.Duration
	InfoURL    *url.URL
	Username   string
	Password   string
}

func parseConfig() (config, error) {
	result := config{
		ListenAddr: ":9205",
		Timeout:    5 * time.Second,
		Username:   os.Getenv("NEXTCLOUD_USERNAME"),
		Password:   os.Getenv("NEXTCLOUD_PASSWORD"),
	}

	rawURL := os.Getenv("NEXTCLOUD_SERVERINFO_URL");
	pflag.StringVarP(&result.ListenAddr, "addr", "a", result.ListenAddr, "Address to listen on for connections.")
	pflag.DurationVarP(&result.Timeout, "timeout", "t", result.Timeout, "Timeout for getting server info document.")
	pflag.StringVarP(&rawURL, "url", "l", rawURL, "URL to nextcloud serverinfo page.")
	pflag.StringVarP(&result.Username, "username", "u", result.Username, "Username for connecting to nextcloud.")
	pflag.StringVarP(&result.Password, "password", "p", result.Password, "Password for connecting to nextcloud.")
	pflag.Parse()

	if len(rawURL) == 0 {
		return result, errors.New("need to provide an info URL")
	}

	infoURL, err := url.Parse(rawURL)
	if err != nil {
		return result, fmt.Errorf("info URL is not valid: %s", err)
	}
	result.InfoURL = infoURL

	if len(result.Username) == 0 {
		return result, errors.New("need to provide a username")
	}

	if len(result.Password) == 0 {
		return result, errors.New("need to provide a password")
	}

	return result, nil
}

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
