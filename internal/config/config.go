package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

const (
	envPrefix        = "NEXTCLOUD_"
	envListenAddress = envPrefix + "LISTEN_ADDRESS"
	envTimeout       = envPrefix + "TIMEOUT"
	envServerURL     = envPrefix + "SERVER"
	envUsername      = envPrefix + "USERNAME"
	envPassword      = envPrefix + "PASSWORD"
)

// Config contains the configuration options for nextcloud-exporter.
type Config struct {
	ListenAddr string        `yaml:"listenAddress"`
	Timeout    time.Duration `yaml:"timeout"`
	ServerURL  string        `yaml:"server"`
	Username   string        `yaml:"username"`
	Password   string        `yaml:"password"`
}

// Get loads the configuration. Flags, environment variables and configuration file are considered.
func Get() (Config, error) {
	return parseConfig(os.Args, os.Getenv)
}

func parseConfig(args []string, envFunc func(string) string) (Config, error) {
	result, configFile, err := loadConfigFromFlags(args)
	if err != nil {
		return Config{}, fmt.Errorf("error parsing flags: %s", err)
	}

	if configFile != "" {
		rawFile, err := loadConfigFromFile(configFile)
		if err != nil {
			return Config{}, fmt.Errorf("error reading configuration file: %s", err)
		}

		result = mergeConfig(result, rawFile)
	}

	env, err := loadConfigFromEnv(envFunc)
	if err != nil {
		return Config{}, fmt.Errorf("error reading environment variables: %s", err)
	}
	result = mergeConfig(result, env)

	if strings.HasPrefix(result.Password, "@") {
		fileName := strings.TrimPrefix(result.Password, "@")
		password, err := readPasswordFile(fileName)
		if err != nil {
			return Config{}, fmt.Errorf("can not read password file: %s", err)
		}

		result.Password = password
	}

	if len(result.ServerURL) == 0 {
		return Config{}, errors.New("need to set a server URL")
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

func loadConfigFromFlags(args []string) (result Config, configFile string, err error) {
	defaults := defaultConfig()

	flags := pflag.NewFlagSet(args[0], pflag.ContinueOnError)
	flags.StringVarP(&configFile, "config-file", "c", "", "Path to YAML configuration file.")
	flags.StringVarP(&result.ListenAddr, "addr", "a", defaults.ListenAddr, "Address to listen on for connections.")
	flags.DurationVarP(&result.Timeout, "timeout", "t", defaults.Timeout, "Timeout for getting server info document.")
	flags.StringVarP(&result.ServerURL, "server", "s", "", "URL to Nextcloud server.")
	flags.StringVarP(&result.Username, "username", "u", defaults.Username, "Username for connecting to Nextcloud.")
	flags.StringVarP(&result.Password, "password", "p", defaults.Password, "Password for connecting to Nextcloud.")

	if err := flags.Parse(args[1:]); err != nil {
		return Config{}, "", err
	}

	return result, configFile, nil
}

func loadConfigFromFile(fileName string) (Config, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return Config{}, err
	}

	var result Config
	if err := yaml.NewDecoder(file).Decode(&result); err != nil {
		return Config{}, err
	}

	return result, nil
}

func loadConfigFromEnv(getEnv func(string) string) (Config, error) {
	result := Config{
		ListenAddr: getEnv(envListenAddress),
		ServerURL:  getEnv(envServerURL),
		Username:   getEnv(envUsername),
		Password:   getEnv(envPassword),
	}

	if raw := getEnv(envTimeout); raw != "" {
		value, err := time.ParseDuration(raw)
		if err != nil {
			return Config{}, err
		}

		result.Timeout = value
	}

	return result, nil
}

func mergeConfig(base, override Config) Config {
	result := base
	if override.ListenAddr != "" {
		result.ListenAddr = override.ListenAddr
	}

	if override.ServerURL != "" {
		result.ServerURL = override.ServerURL
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

	return strings.TrimSuffix(string(bytes), "\n"), nil
}
