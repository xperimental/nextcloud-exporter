package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

type config struct {
	ListenAddr string
	Timeout    time.Duration
	InfoURL    *url.URL
	Username   string
	Password   string
}

type rawConfig struct {
	ListenAddr string        `yaml:"listenAddress"`
	Timeout    time.Duration `yaml:"timeout"`
	InfoURL    string        `yaml:"infoUrl"`
	Username   string        `yaml:"username"`
	Password   string        `yaml:"password"`
}

func parseConfig() (config, error) {
	raw := loadConfigFromFlags()

	if len(raw.InfoURL) == 0 {
		return config{}, errors.New("need to provide an info URL")
	}

	result, err := convertConfig(raw)
	if err != nil {
		return config{}, err
	}

	if len(result.Username) == 0 {
		return config{}, errors.New("need to provide a username")
	}

	if len(result.Password) == 0 {
		return config{}, errors.New("need to provide a password")
	}

	return result, nil
}

func defaultConfig() config {
	return config{
		ListenAddr: ":9205",
		Timeout:    5 * time.Second,
	}
}

func loadConfigFromFlags() (result rawConfig) {
	defaults := defaultConfig()
	pflag.StringVarP(&result.ListenAddr, "addr", "a", defaults.ListenAddr, "Address to listen on for connections.")
	pflag.DurationVarP(&result.Timeout, "timeout", "t", defaults.Timeout, "Timeout for getting server info document.")
	pflag.StringVarP(&result.InfoURL, "url", "l", "", "URL to nextcloud serverinfo page.")
	pflag.StringVarP(&result.Username, "username", "u", defaults.Username, "Username for connecting to nextcloud.")
	pflag.StringVarP(&result.Password, "password", "p", defaults.Password, "Password for connecting to nextcloud.")
	pflag.Parse()

	return result
}

func convertConfig(raw rawConfig) (config, error) {
	result := config{
		ListenAddr: raw.ListenAddr,
		Timeout:    raw.Timeout,
		Username:   raw.Username,
		Password:   raw.Password,
	}

	infoURL, err := url.Parse(raw.InfoURL)
	if err != nil {
		return config{}, fmt.Errorf("info URL is not valid: %s", err)
	}
	result.InfoURL = infoURL

	if strings.HasPrefix(result.Password, "@") {
		fileName := strings.TrimPrefix(result.Password, "@")
		password, err := readPasswordFile(fileName)
		if err != nil {
			return config{}, fmt.Errorf("can not read password file: %s", err)
		}

		result.Password = password
	}

	return result, nil
}

func readPasswordFile(fileName string) (string, error) {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
