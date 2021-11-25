package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

func RegisterInfoMetric(version, gitCommit string) error {
	infoMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: metricPrefix + "exporter_info",
		Help: "Information about the nextcloud-exporter.",
		ConstLabels: prometheus.Labels{
			"version": version,
			"commit":  gitCommit,
		},
	})
	infoMetric.Set(1)

	return prometheus.Register(infoMetric)
}
