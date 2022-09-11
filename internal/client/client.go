package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/xperimental/nextcloud-exporter/serverinfo"
)

const (
	nextcloudTokenHeader = "NC-Token"
)

var (
	ErrNotAuthorized = errors.New("wrong credentials")
	ErrRatelimit     = errors.New("too many requests")
)

type InfoClient func() (*serverinfo.ServerInfo, error)

func New(infoURL, username, password, authToken string, timeout time.Duration, userAgent string, tlsSkipVerify bool) InfoClient {
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				// disable TLS certification verification, if desired
				InsecureSkipVerify: tlsSkipVerify,
			},
		},
	}

	return func() (*serverinfo.ServerInfo, error) {
		req, err := http.NewRequest(http.MethodGet, infoURL, nil)
		if err != nil {
			return nil, err
		}

		if authToken == "" {
			req.SetBasicAuth(username, password)
		} else {
			req.Header.Set(nextcloudTokenHeader, authToken)
		}

		req.Header.Set("User-Agent", userAgent)

		res, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		switch res.StatusCode {
		case http.StatusOK:
			break
		case http.StatusUnauthorized:
			return nil, ErrNotAuthorized
		case http.StatusTooManyRequests:
			return nil, ErrRatelimit
		default:
			return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
		}

		status, err := serverinfo.ParseJSON(res.Body)
		if err != nil {
			return nil, fmt.Errorf("can not parse server info: %w", err)
		}

		return status, nil
	}
}
