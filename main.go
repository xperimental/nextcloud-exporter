package main

import (
	"bufio"
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
	ListenAddr   string
	Timeout      time.Duration
	InfoURL      *url.URL
	Username     string
	Password     string
	PasswordFile string
}

func parseConfig() (config, error) {
	result := config{
		ListenAddr:   ":9205",
		Timeout:      5 * time.Second,
		Username:     os.Getenv("NEXTCLOUD_USERNAME"),
		Password:     os.Getenv("NEXTCLOUD_PASSWORD"),
		PasswordFile: os.Getenv("NEXTCLOUD_PASSWORD_FILE"),
	}

	rawURL := os.Getenv("NEXTCLOUD_SERVERINFO_URL");
	pflag.StringVarP(&result.ListenAddr, "addr", "a", result.ListenAddr, "Address to listen on for connections.")
	pflag.DurationVarP(&result.Timeout, "timeout", "t", result.Timeout, "Timeout for getting server info document.")
	pflag.StringVarP(&rawURL, "url", "l", rawURL, "URL to nextcloud serverinfo page.")
	pflag.StringVarP(&result.Username, "username", "u", result.Username, "Username for connecting to nextcloud.")
	pflag.StringVarP(&result.Password, "password", "p", result.Password, "Password for connecting to nextcloud.")
	pflag.StringVar(&result.PasswordFile, "password-file", result.PasswordFile, "File containing the password for connecting to nextcloud.")
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

	if len(result.PasswordFile) != 0 {
		fd, err := os.Open(result.PasswordFile)
		if err != nil {
			return result, fmt.Errorf("could not open password file: %s", err)
		}
		defer fd.Close()

		s := bufio.NewScanner(fd)
		if s.Scan() {
			result.Password = s.Text()
			return result, nil
		} else {
			return result, errors.New("read empty password from given password file")
		}
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
