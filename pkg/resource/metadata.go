package resource

import "github.com/tkeel-io/core/pkg/config"

type Metadata struct {
	Name       string
	Properties map[string]string `json:"properties"`
}

func ParseFrom(meta *config.Metadata) Metadata {
	m := Metadata{Name: meta.Name}
	if len(meta.Properties) > 0 {
		m.Properties = make(map[string]string)
		for _, pair := range meta.Properties {
			m.Properties[pair.Key] = pair.Value
		}
	}
	return m
}
