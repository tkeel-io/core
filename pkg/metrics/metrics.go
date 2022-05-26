package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// metrics label.
	MetricsLabelTenant = "tenant_id"

	// metrics rawdata count name.
	MetricsRawDataCount = "rawdata_lines"

	// metrics rawdata storage name.
	MetricsRawDataStorgae = "rawdata_storage"

	// metrics timeseries count name.
	MetricsTimeseriesCount = "timeseries_lines"

	// metrics rawdata storage name.
	MetricsTimeseriesStorgae = "timeseries_storage"
)

var CollectorRawDataCount = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: MetricsRawDataCount,
		Help: "rawdata count.",
	},
	[]string{MetricsLabelTenant},
)

var CollectorRawDataStorage = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: MetricsRawDataStorgae,
		Help: "rawdata storage.",
	},
	[]string{MetricsLabelTenant},
)

var CollectorTimeseriesCount = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: MetricsTimeseriesCount,
		Help: "timeseries count.",
	},
	[]string{MetricsLabelTenant},
)

var CollectorTimeseriesStorage = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: MetricsTimeseriesStorgae,
		Help: "timeseries storage.",
	},
	[]string{MetricsLabelTenant},
)

var Metrics = []prometheus.Collector{CollectorRawDataCount, CollectorRawDataStorage, CollectorTimeseriesCount, CollectorTimeseriesStorage}
