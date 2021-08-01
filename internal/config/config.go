package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
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
	envTLSSkipVerify = envPrefix + "TLS_SKIP_VERIFY"
)

// RunMode signals what the main application should do after parsing the options.
type RunMode int

const (
	// RunModeExporter is the normal operation as an exporter serving metrics via HTTP.
	RunModeExporter RunMode = iota
	// RunModeHelp shows information about available options.
	RunModeHelp
	// RunModeLogin is used to interactively login to a Nextcloud instance.
	RunModeLogin
	// RunModeVersion shows version information.
	RunModeVersion
)

func (m RunMode) String() string {
	switch m {
	case RunModeExporter:
		return "exporter"
	case RunModeHelp:
		return "help"
	case RunModeLogin:
		return "login"
	case RunModeVersion:
		return "version"
	default:
		return "error"
	}
}

// Config contains the configuration options for nextcloud-exporter.
type Config struct {
	ListenAddr    string        `yaml:"listenAddress"`
	Timeout       time.Duration `yaml:"timeout"`
	ServerURL     string        `yaml:"server"`
	Username      string        `yaml:"username"`
	Password      string        `yaml:"password"`
	TLSSkipVerify bool          `yaml:"tlsSkipVerify"`
	UseJSON       bool          `yaml:"json"`
	RunMode       RunMode
}

// Validate checks if the configuration contains all necessary parameters.
func (c Config) Validate() error {
	if len(c.ServerURL) == 0 {
		return errors.New("need to set a server URL")
	}

	if len(c.Username) == 0 {
		return errors.New("need to provide a username")
	}

	if len(c.Password) == 0 {
		return errors.New("need to provide a password")
	}

	return nil
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
	flags.BoolVar(&result.TLSSkipVerify, "tls-skip-verify", defaults.TLSSkipVerify, "Skip certificate verification of Nextcloud server.")
	flags.BoolVar(&result.UseJSON, "use-json", defaults.UseJSON, "Read data in JSON format from Nextcloud server.")
	modeLogin := flags.Bool("login", false, "Use interactive login to create app password.")
	modeVersion := flags.BoolP("version", "V", false, "Show version information and exit.")

	if err := flags.Parse(args[1:]); err != nil {
		if err == pflag.ErrHelp {
			return Config{
				RunMode: RunModeHelp,
			}, "", nil
		}

		return Config{}, "", err
	}

	if *modeVersion {
		return Config{
			RunMode: RunModeVersion,
		}, "", nil
	}

	if *modeLogin {
		result.RunMode = RunModeLogin
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
	tlsSkipVerify := false
	if rawValue := getEnv(envTLSSkipVerify); rawValue != "" {
		value, err := strconv.ParseBool(rawValue)
		if err != nil {
			return Config{}, fmt.Errorf("can not parse value for %q: %s", envTLSSkipVerify, rawValue)
		}
		tlsSkipVerify = value
	}

	result := Config{
		ListenAddr:    getEnv(envListenAddress),
		ServerURL:     getEnv(envServerURL),
		Username:      getEnv(envUsername),
		Password:      getEnv(envPassword),
		TLSSkipVerify: tlsSkipVerify,
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

	if override.TLSSkipVerify {
		result.TLSSkipVerify = override.TLSSkipVerify
	}

	if override.UseJSON {
		result.UseJSON = override.UseJSON
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
