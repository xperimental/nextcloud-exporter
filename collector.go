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

const (
	metricPrefix = "nextcloud_"
)

var (
	systemInfoDesc = prometheus.NewDesc(
		metricPrefix+"system_info",
		"Contains meta information about Nextcloud as labels. Value is always 1.",
		[]string{"version"}, nil)
	usersDesc = prometheus.NewDesc(
		metricPrefix+"users_total",
		"Number of users of the instance.",
		nil, nil)
	filesDesc = prometheus.NewDesc(
		metricPrefix+"files_total",
		"Number of files served by the instance.",
		nil, nil)
	freeSpaceDesc = prometheus.NewDesc(
		metricPrefix+"free_space_bytes",
		"Number of bytes of free space on the instance.",
		nil, nil)
	sharesDesc = prometheus.NewDesc(
		metricPrefix+"shares_total",
		"Number of shares by type.",
		[]string{"type"}, nil)
	federationsDesc = prometheus.NewDesc(
		metricPrefix+"shares_federated_total",
		"Number of federated shares by direction.",
		[]string{"direction"}, nil)
	activeUsersDesc = prometheus.NewDesc(
		metricPrefix+"active_users_total",
		"Number of active users for the last five minutes.",
		nil, nil)
	phpInfoDesc = prometheus.NewDesc(
		metricPrefix+"php_info",
		"Contains meta information about PHP as labels. Value is always 1.",
		[]string{"version"}, nil)
	phpMemoryLimitDesc = prometheus.NewDesc(
		metricPrefix+"php_memory_limit_bytes",
		"Configured PHP memory limit in bytes.",
		nil, nil)
	phpMaxUploadSizeDesc = prometheus.NewDesc(
		metricPrefix+"php_upload_max_size_bytes",
		"Configured maximum upload size in bytes.",
		nil, nil)
	databaseSizeDesc = prometheus.NewDesc(
		metricPrefix+"database_size_bytes",
		"Size of database in bytes as reported from engine.",
		nil, nil)
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
			Name: metricPrefix + "up",
			Help: "Shows if nextcloud is deemed up by the collector.",
		}),
		authErrorsMetric: prometheus.NewCounter(prometheus.CounterOpts{
			Name: metricPrefix + "auth_errors_total",
			Help: "Counts number of authentication errors encountered by the collector.",
		}),
		scrapeErrorsMetric: prometheus.NewCounter(prometheus.CounterOpts{
			Name: metricPrefix + "scrape_errors_total",
			Help: "Counts the number of scrape errors by this collector.",
		}),
	}
}

func (c *nextcloudCollector) Describe(ch chan<- *prometheus.Desc) {
	c.upMetric.Describe(ch)
	c.authErrorsMetric.Describe(ch)
	c.scrapeErrorsMetric.Describe(ch)
	ch <- usersDesc
	ch <- filesDesc
	ch <- freeSpaceDesc
	ch <- sharesDesc
	ch <- federationsDesc
	ch <- activeUsersDesc
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

	if err := collectSimpleMetrics(ch, status); err != nil {
		return err
	}

	if err := collectShares(ch, status.Data.Nextcloud.Shares); err != nil {
		return err
	}

	if err := collectFederatedShares(ch, status.Data.Nextcloud.Shares); err != nil {
		return err
	}

	systemInfo := []string{
		status.Data.Nextcloud.System.Version,
	}
	if err := collectInfoMetric(ch, systemInfoDesc, systemInfo); err != nil {
		return err
	}

	phpInfo := []string{
		status.Data.Server.PHP.Version,
	}
	if err := collectInfoMetric(ch, phpInfoDesc, phpInfo); err != nil {
		return err
	}

	return nil
}

func collectSimpleMetrics(ch chan<- prometheus.Metric, status serverinfo.ServerInfo) error {
	metrics := []struct {
		desc  *prometheus.Desc
		value float64
	}{
		{
			desc:  usersDesc,
			value: float64(status.Data.Nextcloud.Storage.Users),
		},
		{
			desc:  filesDesc,
			value: float64(status.Data.Nextcloud.Storage.Files),
		},
		{
			desc:  freeSpaceDesc,
			value: float64(status.Data.Nextcloud.System.FreeSpace),
		},
		{
			desc:  activeUsersDesc,
			value: float64(status.Data.ActiveUsers.Last5Minutes),
		},
		{
			desc:  phpMemoryLimitDesc,
			value: float64(status.Data.Server.PHP.MemoryLimit),
		},
		{
			desc:  phpMaxUploadSizeDesc,
			value: float64(status.Data.Server.PHP.UploadMaxFilesize),
		},
		{
			desc:  databaseSizeDesc,
			value: float64(status.Data.Server.Database.Size),
		},
	}
	for _, m := range metrics {
		metric, err := prometheus.NewConstMetric(m.desc, prometheus.GaugeValue, m.value)
		if err != nil {
			return fmt.Errorf("error creating metric for %s: %s", m.desc, err)
		}
		ch <- metric
	}

	return nil
}

func collectShares(ch chan<- prometheus.Metric, shares serverinfo.Shares) error {
	values := make(map[string]float64)
	values["user"] = float64(shares.SharesUser)
	values["group"] = float64(shares.SharesGroups)
	values["authlink"] = float64(shares.SharesLink - shares.SharesLinkNoPassword)
	values["link"] = float64(shares.SharesLink)

	return collectMap(ch, sharesDesc, values)
}

func collectFederatedShares(ch chan<- prometheus.Metric, shares serverinfo.Shares) error {
	values := make(map[string]float64)
	values["sent"] = float64(shares.FedSent)
	values["received"] = float64(shares.FedReceived)

	return collectMap(ch, federationsDesc, values)
}

func collectMap(ch chan<- prometheus.Metric, desc *prometheus.Desc, labelValueMap map[string]float64) error {
	for k, v := range labelValueMap {
		metric, err := prometheus.NewConstMetric(desc, prometheus.GaugeValue, v, k)
		if err != nil {
			return fmt.Errorf("error creating shares metric for %s: %s", k, err)
		}
		ch <- metric
	}

	return nil
}

func collectInfoMetric(ch chan<- prometheus.Metric, desc *prometheus.Desc, labelValues []string) error {
	metric, err := prometheus.NewConstMetric(desc, prometheus.GaugeValue, 1, labelValues...)
	if err != nil {
		return err
	}

	ch <- metric
	return nil
}
