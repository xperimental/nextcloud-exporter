package login

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
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

// Login contains the login information gathered during the login session.
type Login struct {
	Username string
	Password string
}

// Client can be used to start an interactive login session with a Nextcloud server.
type Client struct {
	log       logrus.FieldLogger
	userAgent string
	serverURL string

	client    *http.Client
	sleepFunc func()
}

// Init creates a new LoginClient. The session can then be started using StartInteractive.
func Init(log logrus.FieldLogger, userAgent, serverURL string, tlsSkipVerify bool) *Client {
	return &Client{
		log:       log,
		userAgent: userAgent,
		serverURL: serverURL,

		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: tlsSkipVerify,
				},
			},
		},
		sleepFunc: func() { time.Sleep(pollInterval) },
	}
}

// StartInteractive starts an interactive login session for the Nextcloud server and user.
// The end-result of this is an app-password for the exporter which should be used instead of a user password.
func (c *Client) StartInteractive() error {
	version, err := c.getMajorVersion()
	if err != nil {
		return fmt.Errorf("error getting version: %w", err)
	}

	if version < minimumMajorVersion {
		return fmt.Errorf("Nextcloud version too old for login: %d Minimum: %d", version, minimumMajorVersion) //nolint:staticcheck
	}

	info, err := c.getLoginInfo()
	if err != nil {
		return fmt.Errorf("error getting login info: %w", err)
	}
	c.log.Infof("Please open this URL in a browser: %s", info.LoginURL)
	c.log.Infoln("Waiting for login ... (Ctrl-C to abort)")

	login, err := c.pollLogin(info.PollInfo)
	if err != nil {
		return fmt.Errorf("error during poll: %w", err)
	}

	c.log.Infof("Username: %s", login.Username)
	c.log.Infof("Password: %s", login.Password)
	return nil
}

func (c *Client) doRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("can not create request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)

	if body != nil {
		req.Header.Set("Content-Type", contentType)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error connecting: %w", err)
	}

	return res, nil
}

func (c *Client) getMajorVersion() (int, error) {
	statusURL := c.serverURL + statusPath
	res, err := c.doRequest(http.MethodGet, statusURL, nil)
	if err != nil {
		return 0, fmt.Errorf("error connecting: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("non-ok status: %d", res.StatusCode)
	}

	var status struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(res.Body).Decode(&status); err != nil {
		return 0, fmt.Errorf("error decoding status: %w", err)
	}

	tokens := strings.SplitN(status.Version, ".", 2)
	version, err := strconv.Atoi(tokens[0])
	if err != nil {
		return 0, fmt.Errorf("can not parse %q as version: %w", status.Version, err)
	}

	return version, nil
}

func (c *Client) getLoginInfo() (loginInfo, error) {
	loginURL := c.serverURL + loginPath
	res, err := c.doRequest(http.MethodPost, loginURL, nil)
	if err != nil {
		return loginInfo{}, fmt.Errorf("error connecting: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return loginInfo{}, fmt.Errorf("non-ok status: %d", res.StatusCode)
	}

	var result loginInfo
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return loginInfo{}, fmt.Errorf("error decoding login info: %w", err)
	}

	return result, nil
}

func (c *Client) pollLogin(info pollInfo) (Login, error) {
	body := fmt.Sprintf("token=%s", info.Token)
	c.log.Debugf("poll endpoint: %s", info.Endpoint)

	for {
		c.sleepFunc()
		reader := strings.NewReader(body)
		res, err := c.doRequest(http.MethodPost, info.Endpoint, reader)
		if err != nil {
			continue
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			c.log.Debugf("poll status: %d", res.StatusCode)
			continue
		}

		var password passwordInfo
		if err := json.NewDecoder(res.Body).Decode(&password); err != nil {
			return Login{}, fmt.Errorf("error decoding password info: %w", err)
		}

		return Login{
			Username: password.LoginName,
			Password: password.AppPassword,
		}, nil
	}
}
