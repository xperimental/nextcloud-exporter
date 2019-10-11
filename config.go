package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"time"

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
		password, err := readPasswordFile(result.PasswordFile)
		if err != nil {
			return result, fmt.Errorf("can not read password file: %s", err)
		}

		result.Password = password
	}
	if len(result.Password) == 0 {
		return result, errors.New("need to provide a password")
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
