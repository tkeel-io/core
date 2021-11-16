package statem

import "github.com/tkeel-io/core/pkg/mapper"

// PropertyMessage state property message.
type StateMessage struct {
	messageBase

	StateID  string `json:"state_id"`
	Operator string `json:"operator"`
}

func NewPropertyMessage(id string, props map[string][]byte) PropertyMessage {
	return PropertyMessage{
		StateID:    id,
		Properties: props,
	}
}

// PropertyMessage state property message.
type PropertyMessage struct {
	messageBase

	StateID    string            `json:"state_id"`
	Properties map[string][]byte `json:"properties"`
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
