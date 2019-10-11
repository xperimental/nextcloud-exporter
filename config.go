package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

const (
	defaultPath = "/ocs/v2.php/apps/serverinfo/api/v1/info"

	envPrefix        = "NEXTCLOUD_"
	envListenAddress = envPrefix + "LISTEN_ADDRESS"
	envTimeout       = envPrefix + "TIMEOUT"
	envInfoURL       = envPrefix + "SERVERINFO_URL"
	envUsername      = envPrefix + "USERNAME"
	envPassword      = envPrefix + "PASSWORD"
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
	raw, configFile := loadConfigFromFlags()

	if configFile != "" {
		rawFile, err := loadConfigFromFile(configFile)
		if err != nil {
			return config{}, fmt.Errorf("error reading configuration file: %s", err)
		}

		raw = mergeConfig(raw, rawFile)
	}

	env, err := loadConfigFromEnv()
	if err != nil {
		return config{}, fmt.Errorf("error reading environment variables: %s", err)
	}
	raw = mergeConfig(raw, env)

	if len(raw.InfoURL) == 0 {
		return config{}, errors.New("need to set an info URL")
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

func loadConfigFromFlags() (result rawConfig, configFile string) {
	defaults := defaultConfig()
	pflag.StringVarP(&configFile, "config-file", "c", "", "Path to YAML configuration file.")
	pflag.StringVarP(&result.ListenAddr, "addr", "a", defaults.ListenAddr, "Address to listen on for connections.")
	pflag.DurationVarP(&result.Timeout, "timeout", "t", defaults.Timeout, "Timeout for getting server info document.")
	pflag.StringVarP(&result.InfoURL, "url", "l", "", "URL to Nextcloud serverinfo page.")
	pflag.StringVarP(&result.Username, "username", "u", defaults.Username, "Username for connecting to Nextcloud.")
	pflag.StringVarP(&result.Password, "password", "p", defaults.Password, "Password for connecting to Nextcloud.")
	pflag.Parse()

	return result, configFile
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

	if result.InfoURL.Path == "" {
		result.InfoURL.Path = defaultPath
	}

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

func loadConfigFromFile(fileName string) (rawConfig, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return rawConfig{}, err
	}

	var result rawConfig
	if err := yaml.NewDecoder(file).Decode(&result); err != nil {
		return rawConfig{}, err
	}

	return result, nil
}

func loadConfigFromEnv() (rawConfig, error) {
	result := rawConfig{
		ListenAddr: os.Getenv(envListenAddress),
		InfoURL:    os.Getenv(envInfoURL),
		Username:   os.Getenv(envUsername),
		Password:   os.Getenv(envPassword),
	}

	if raw, ok := os.LookupEnv(envTimeout); ok {
		value, err := time.ParseDuration(raw)
		if err != nil {
			return rawConfig{}, err
		}

		result.Timeout = value
	}

	return result, nil
}

func mergeConfig(base, override rawConfig) rawConfig {
	result := base
	if override.ListenAddr != "" {
		result.ListenAddr = override.ListenAddr
	}

	if override.InfoURL != "" {
		result.InfoURL = override.InfoURL
	}

	if override.Username != "" {
		result.Username = override.Username
	}

	if override.Password != "" {
		result.Password = override.Password
	}

	if override.Timeout != 0 {
		result.Timeout = override.Timeout
	}

	return result
}

func readPasswordFile(fileName string) (string, error) {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
