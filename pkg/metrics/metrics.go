package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// metrics label.
	MetricsLabelTenant    = "tenant_id"
	MetricsLabelMsgType   = "msg_type"
	MetricsLabelSpaceType = "space_type"

	// msg type.
	MsgTypeSubscribe  = "subscribe"
	MsgTypeRawData    = "rawdata"
	MsgTypeTimeseries = "timeseries"

	// space type.
	SpaceTypeTotal = "total"
	SpaceTypeUsed  = "used"

	// metrics msg count name.
	MetricsMsgCount = "core_msg_total"

	// metrics rawdata storage name.
	MetricsRawDataStorgae = "rawdata_storage"

	// metrics rawdata storage name.
	MetricsTimeseriesStorgae = "timeseries_storage"

	// metrics message storage name.
	MetricsMsgStorageSpace = "msg_storage_space"

	// metrics message storage days.
	MetricsMsgStorageSeconds = "msg_storage_seconds"
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

var CollectorMsgStorageSpace = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: MetricsMsgStorageSpace,
		Help: "msg storage space.",
	},
	[]string{MetricsLabelTenant, MetricsLabelSpaceType},
)

var CollectorMsgStorageSeconds = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: MetricsMsgStorageSeconds,
		Help: "msg storage seconds.",
	},
	[]string{MetricsLabelTenant},
)

var Metrics = []prometheus.Collector{CollectorRawDataStorage, CollectorTimeseriesStorage, CollectorMsgCount, CollectorMsgStorageSpace, CollectorMsgStorageSeconds}
