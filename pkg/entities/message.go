package entities

import "github.com/tkeel-io/core/pkg/mapper"

// EntityAsyncMessage entity message.
type EntityMessage struct {
	messageBase

	SourceID       string                 `json:"source_id"`
	Values         map[string]interface{} `json:"values"`
	PromiseHandler PromiseFunc            `json:"-"`
}

func (esm EntityMessage) Promise() PromiseFunc { return esm.PromiseHandler }

type TentacleMsg struct {
	messageBase

	Operator string            `json:"operator"`
	TargetID string            `json:"target"` //nolint
	Items    []mapper.WatchKey `json:"items"`
}
