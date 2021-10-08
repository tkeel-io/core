package service

import (
	"context"
	"errors"

	"github.com/dapr/go-sdk/service/common"
)

// TimeSeriesService is a time-series service.
type TimeSeriesService struct {
}

// NewTimeSeriesService returns a new TimeSeriesService
func NewTimeSeriesService() *TimeSeriesService {
	return &TimeSeriesService{}
}

// Name return the name.
func (this *TimeSeriesService) Name() string {
	return "time_series"
}

// RegisterService register some method
func (this *TimeSeriesService) RegisterService(daprService common.Service) error {
	//register all handlers.
	if err := daprService.AddServiceInvocationHandler("echo", this.Echo); nil != err {
		log.Error("add service handler failed.", err)
		return err
	}
	return nil
}

// Echo test for RegisterService.
func (this *TimeSeriesService) Echo(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {

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
