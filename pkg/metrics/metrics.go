package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// metrics label.
	MetricsLabelTenant  = "tenant"
	MetricsLabelMsgType = "msg_type"

	// msg type.
	MsgTypeSubscribe  = "subscribe"
	MsgTypeRawData    = "rawdata"
	MsgTypeTimeseries = "timeseries"

	// metrics msg count name.
	MetricsMsgCount = "core_msg_total"

	// metrics rawdata storage name.
	MetricsRawDataStorgae = "rawdata_storage"

	// metrics rawdata storage name.
	MetricsTimeseriesStorgae = "timeseries_storage"
)

var CollectorMsgCount = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: MetricsMsgCount,
		Help: "msg count.",
	},
	[]string{MetricsLabelTenant, MetricsLabelMsgType},
)

var CollectorRawDataStorage = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: MetricsRawDataStorgae,
		Help: "rawdata storage.",
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

var Metrics = []prometheus.Collector{CollectorRawDataStorage, CollectorTimeseriesStorage, CollectorMsgCount}
