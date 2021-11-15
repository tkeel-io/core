package statem

import "github.com/tkeel-io/core/pkg/mapper"

func NewPropertyMessage(id string, props map[string]interface{}) PropertyMessage {
	return PropertyMessage{
		StateID:    id,
		Properties: props,
	}
}

// PropertyMessage state property message.
type PropertyMessage struct {
	messageBase

	StateID    string                 `json:"state_id"`
	Properties map[string]interface{} `json:"properties"`
}

func (esm PropertyMessage) Promise() PromiseFunc { return esm.PromiseHandler }

type MapperMessage struct {
	messageBase

	Operator string     `json:"operator"`
	Mapper   MapperDesc `json:"mapper"`
}

type TentacleMsg struct {
	messageBase

	Operator string            `json:"operator"`
	StateID  string            `json:"state_id"`
	Items    []mapper.WatchKey `json:"items"`
}
