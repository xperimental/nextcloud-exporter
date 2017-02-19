package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/xperimental/nextcloud-exporter/serverinfo"
)

type nextcloudCollector struct {
	infoURL            *url.URL
	username           string
	password           string
	client             *http.Client
	upMetric           prometheus.Gauge
	authErrorsMetric   prometheus.Counter
	scrapeErrorsMetric prometheus.Counter
}

func newCollector(infoURL *url.URL, username, password string, timeout time.Duration) *nextcloudCollector {
	return &nextcloudCollector{
		infoURL:  infoURL,
		username: username,
		password: password,
		client: &http.Client{
			Timeout: timeout,
		},
		upMetric: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nextcloud_up",
			Help: "Shows if nextcloud is deemed up by the collector.",
		}),
		authErrorsMetric: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "nextcloud_auth_errors_total",
			Help: "Counts number of authentication errors encountered by the collector.",
		}),
		scrapeErrorsMetric: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "nextcloud_scrape_errors_total",
			Help: "Counts the number of scrape errors by this collector.",
		}),
	}
}

func (c *nextcloudCollector) Describe(ch chan<- *prometheus.Desc) {
	c.upMetric.Describe(ch)
	c.authErrorsMetric.Describe(ch)
	c.scrapeErrorsMetric.Describe(ch)
}

func (c *nextcloudCollector) Collect(ch chan<- prometheus.Metric) {
	if err := c.collectNextcloud(ch); err != nil {
		log.Printf("Error during scrape: %s", err)

		c.scrapeErrorsMetric.Inc()
		c.upMetric.Set(0)
	} else {
		c.upMetric.Set(1)
	}

	c.upMetric.Collect(ch)
	c.authErrorsMetric.Collect(ch)
	c.scrapeErrorsMetric.Collect(ch)
}

func (c *nextcloudCollector) collectNextcloud(ch chan<- prometheus.Metric) error {
	req, err := http.NewRequest(http.MethodGet, c.infoURL.String(), nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.username, c.password)

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		c.authErrorsMetric.Inc()
		return fmt.Errorf("wrong credentials")
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var status serverinfo.ServerInfo
	if err := xml.NewDecoder(res.Body).Decode(&status); err != nil {
		return err
	}

	return nil
}
