package service

import (
	"net/http"
	"time"

	go_restful "github.com/emicklei/go-restful"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tkeel-io/core/pkg/config"
	logf "github.com/tkeel-io/core/pkg/logfield"
	"github.com/tkeel-io/core/pkg/metrics"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/rawdata"
	"github.com/tkeel-io/core/pkg/resource/tseries"
	"github.com/tkeel-io/kit/log"
)

type MetricsService struct {
	MetricsHandler http.Handler
	rawdataClient  rawdata.Service
	tseriesClient  tseries.TimeSerier
}

func NewMetricsService(mtrCollectors ...prometheus.Collector) (*MetricsService, error) {
	// Create a new registry.
	reg := prometheus.NewRegistry()
	reg.MustRegister(mtrCollectors...)

	metricHandler := promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			EnableOpenMetrics: false,
		},
	)
	tseriesClient := tseries.NewTimeSerier(config.Get().Components.TimeSeries.Name)
	if err := tseriesClient.Init(resource.ParseFrom(config.Get().Components.TimeSeries)); nil != err {
		log.L().Error("initialize time series", logf.Error(err))
		return nil, errors.Wrap(err, "init ts service")
	}
	rawdataClient := rawdata.NewRawDataService(config.Get().Components.Rawdata.Name)
	if err := rawdataClient.Init(resource.ParseFrom(config.Get().Components.Rawdata)); err != nil {
		log.L().Error("initialize rawdata server", logf.Error(err))
	}

	svc := &MetricsService{metricHandler, rawdataClient, tseriesClient}
	go func() {
		svc.flushMetrics()
		timer := time.NewTicker(time.Hour)
		for range timer.C {
			svc.flushMetrics()
		}
	}()
	return svc, nil
}

func (svc *MetricsService) Metrics(req *go_restful.Request, resp *go_restful.Response) {
	svc.MetricsHandler.ServeHTTP(resp, req.Request)
}

func (svc *MetricsService) flushMetrics() {
	_, storage := svc.tseriesClient.GetMetrics()
	metrics.CollectorTimeseriesStorage.WithLabelValues("admin").Set(storage)

	_, storage = svc.rawdataClient.GetMetrics()
	metrics.CollectorRawDataStorage.WithLabelValues("admin").Set(storage)
}
