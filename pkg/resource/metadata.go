package resource

import (
	"github.com/tkeel-io/core/pkg/config"
)

type TimeSeriesMetadata struct {
	Name       string
	Properties map[string]string `json:"properties"`
}

func GetTimeSeriesMetadata(name string) TimeSeriesMetadata {
	m := TimeSeriesMetadata{Name: name}
	for i := 0; i < len(config.Get().TimeSeries); i++ {
		if config.Get().TimeSeries[i].Name != name {
			continue
		}
		if len(config.Get().TimeSeries[i].Properties) > 0 {
			m.Properties = make(map[string]string)
			for _, pair := range config.Get().TimeSeries[i].Properties {
				m.Properties[pair.Key] = pair.Value
			}
		}
		return m
	}
	return m
}
