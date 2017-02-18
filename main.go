package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"
)

type config struct {
	ListenAddr string
	InfoURL    *url.URL
	Username   string
	Password   string
}

func parseConfig() (config, error) {
	result := config{
		ListenAddr: ":8080",
	}

	var rawURL string
	pflag.StringVarP(&result.ListenAddr, "addr", "a", result.ListenAddr, "Address to listen on for connections.")
	pflag.StringVarP(&rawURL, "url", "l", "", "URL to nextcloud serverinfo page.")
	pflag.StringVarP(&result.Username, "username", "u", "", "Username for connecting to nextcloud.")
	pflag.StringVarP(&result.Password, "password", "p", "", "Password for connecting to nextcloud.")
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
	collector := newCollector(config.InfoURL, config.Username, config.Password)
	if err := prometheus.Register(collector); err != nil {
		log.Fatalf("Failed to register collector: %s", err)
	}

	http.Handle("/metrics", prometheus.UninstrumentedHandler())
	http.Handle("/", http.RedirectHandler("/metrics", http.StatusFound))

	log.Printf("Listen on %s...", config.ListenAddr)
	log.Fatal(http.ListenAndServe(config.ListenAddr, nil))
}
