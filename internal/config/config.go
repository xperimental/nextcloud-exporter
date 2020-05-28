package config

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

// Config contains the configuration options for nextcloud-exporter.
type Config struct {
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

// Get loads the configuration. Flags, environment variables and configuration file are considered.
func Get() (Config, error) {
	return parseConfig(os.Args, os.Getenv)
}

func parseConfig(args []string, envFunc func(string) string) (Config, error) {
	raw, configFile, err := loadConfigFromFlags(args)
	if err != nil {
		return Config{}, err
	}

	if configFile != "" {
		rawFile, err := loadConfigFromFile(configFile)
		if err != nil {
			return Config{}, fmt.Errorf("error reading configuration file: %s", err)
		}

		raw = mergeConfig(raw, rawFile)
	}

	env, err := loadConfigFromEnv(envFunc)
	if err != nil {
		return Config{}, fmt.Errorf("error reading environment variables: %s", err)
	}
	raw = mergeConfig(raw, env)

	if len(raw.InfoURL) == 0 {
		return Config{}, errors.New("need to set an info URL")
	}

	result, err := convertConfig(raw)
	if err != nil {
		return Config{}, err
	}

	if len(result.Username) == 0 {
		return Config{}, errors.New("need to provide a username")
	}

	if len(result.Password) == 0 {
		return Config{}, errors.New("need to provide a password")
	}

	return result, nil
}

func defaultConfig() Config {
	return Config{
		ListenAddr: ":9205",
		Timeout:    5 * time.Second,
	}
}

func loadConfigFromFlags(args []string) (result rawConfig, configFile string, err error) {
	defaults := defaultConfig()

	flags := pflag.NewFlagSet(args[0], pflag.ContinueOnError)
	flags.StringVarP(&configFile, "config-file", "c", "", "Path to YAML configuration file.")
	flags.StringVarP(&result.ListenAddr, "addr", "a", defaults.ListenAddr, "Address to listen on for connections.")
	flags.DurationVarP(&result.Timeout, "timeout", "t", defaults.Timeout, "Timeout for getting server info document.")
	flags.StringVarP(&result.InfoURL, "url", "l", "", "URL to Nextcloud serverinfo page.")
	flags.StringVarP(&result.Username, "username", "u", defaults.Username, "Username for connecting to Nextcloud.")
	flags.StringVarP(&result.Password, "password", "p", defaults.Password, "Password for connecting to Nextcloud.")

	if err := flags.Parse(args[1:]); err != nil {
		return rawConfig{}, "", err
	}

	return result, configFile, nil
}

func convertConfig(raw rawConfig) (Config, error) {
	result := Config{
		ListenAddr: raw.ListenAddr,
		Timeout:    raw.Timeout,
		Username:   raw.Username,
		Password:   raw.Password,
	}

	infoURL, err := url.Parse(raw.InfoURL)
	if err != nil {
		return Config{}, fmt.Errorf("info URL is not valid: %s", err)
	}
	result.InfoURL = infoURL

	if result.InfoURL.Path == "" {
		result.InfoURL.Path = defaultPath
	}

	if strings.HasPrefix(result.Password, "@") {
		fileName := strings.TrimPrefix(result.Password, "@")
		password, err := readPasswordFile(fileName)
		if err != nil {
			return Config{}, fmt.Errorf("can not read password file: %s", err)
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

func loadConfigFromEnv(getEnv func(string) string) (rawConfig, error) {
	result := rawConfig{
		ListenAddr: getEnv(envListenAddress),
		InfoURL:    getEnv(envInfoURL),
		Username:   getEnv(envUsername),
		Password:   getEnv(envPassword),
	}

	if raw := getEnv(envTimeout); raw != "" {
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
