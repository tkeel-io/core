package service

import (
	"context"

	"github.com/dapr/go-sdk/service/common"
	"github.com/pkg/errors"
)

// TimeSeriesService is a time-series service.
type TimeSeriesService struct {
}

// NewTimeSeriesService returns a new TimeSeriesService.
func NewTimeSeriesService() *TimeSeriesService {
	return &TimeSeriesService{}
}

// Name return the name.
func (s *TimeSeriesService) Name() string {
	return "time_series"
}

// RegisterService register some method.
func (s *TimeSeriesService) RegisterService(daprService common.Service) error {
	// register all handlers.
	if err := daprService.AddServiceInvocationHandler("echo", s.Echo); nil != err {
		log.Error("add service handler failed.", err)
		return errors.Wrap(err, "dapr service 'echo' invocation handler err")
	}
	return nil
}

// Echo test for RegisterService.
func (s *TimeSeriesService) Echo(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	if in == nil {
		err = errors.New("nil invocation parameter")
		return
	}

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}
	return
}
