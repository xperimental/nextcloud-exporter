package login

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	statusPath          = "/status.php"
	minimumMajorVersion = 16

	loginPath = "/index.php/login/v2"

	pollInterval = time.Second
	contentType  = "application/x-www-form-urlencoded"
)

var (
	client = &http.Client{
		Timeout: 30 * time.Second,
	}
)

type loginInfo struct {
	LoginURL string   `json:"login"`
	PollInfo pollInfo `json:"poll"`
}

type pollInfo struct {
	Token    string `json:"token"`
	Endpoint string `json:"endpoint"`
}

type passwordInfo struct {
	Server      string `json:"server"`
	LoginName   string `json:"loginName"`
	AppPassword string `json:"appPassword"`
}

// StartInteractive starts an interactive login session for the Nextcloud server and user.
// The end-result of this is an app-password for the exporter which should be used instead of a user password.
func StartInteractive(log logrus.FieldLogger, userAgent, serverURL, username string) error {
	version, err := getMajorVersion(serverURL)
	if err != nil {
		return fmt.Errorf("error getting version: %s", err)
	}

	if version < minimumMajorVersion {
		return fmt.Errorf("Nextcloud version too old for login: %d Minimum: %d", version, minimumMajorVersion)
	}

	info, err := getLoginInfo(userAgent, serverURL)
	if err != nil {
		return fmt.Errorf("error getting login info: %s", err)
	}
	log.Infof("Please open this URL in a browser: %s", info.LoginURL)
	log.Infoln("Waiting for login ...")

	password, err := pollPassword(log, userAgent, info.PollInfo)
	if err != nil {
		return fmt.Errorf("error during poll: %s", err)
	}

	log.Infof("Your app password is: %s", password)
	return nil
}

func getMajorVersion(serverURL string) (int, error) {
	statusURL := serverURL + statusPath
	res, err := client.Get(statusURL)
	if err != nil {
		return 0, fmt.Errorf("error connecting: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("non-ok status: %d", res.StatusCode)
	}

	var status struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(res.Body).Decode(&status); err != nil {
		return 0, fmt.Errorf("error decoding status: %s", err)
	}

	tokens := strings.SplitN(status.Version, ".", 2)
	version, err := strconv.Atoi(tokens[0])
	if err != nil {
		return 0, fmt.Errorf("can not parse %q as version: %s", status.Version, err)
	}

	return version, nil
}

func getLoginInfo(userAgent, serverURL string) (loginInfo, error) {
	loginURL := serverURL + loginPath
	req, err := http.NewRequest(http.MethodPost, loginURL, nil)
	if err != nil {
		return loginInfo{}, fmt.Errorf("can not create request: %s", err)
	}
	req.Header.Set("User-Agent", userAgent)

	res, err := client.Do(req)
	if err != nil {
		return loginInfo{}, fmt.Errorf("error connecting: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return loginInfo{}, fmt.Errorf("non-ok status: %d", res.StatusCode)
	}

	var result loginInfo
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return loginInfo{}, fmt.Errorf("error parsing login info: %s", err)
	}

	return result, nil
}

func pollPassword(log logrus.FieldLogger, userAgent string, info pollInfo) (string, error) {
	body := fmt.Sprintf("token=%s", info.Token)
	log.Debugf("poll endpoint: %s", info.Endpoint)

	for {
		time.Sleep(pollInterval)
		reader := strings.NewReader(body)
		req, err := http.NewRequest(http.MethodPost, info.Endpoint, reader)
		if err != nil {
			return "", fmt.Errorf("can not create request: %s", err)
		}
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Content-Type", contentType)

		res, err := client.Do(req)
		if err != nil {
			continue
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			log.Debugf("poll status: %d", res.StatusCode)
			continue
		}

		var password passwordInfo
		if err := json.NewDecoder(res.Body).Decode(&password); err != nil {
			return "", fmt.Errorf("error decoding password info: %s", err)
		}

		return password.AppPassword, nil
	}
}
