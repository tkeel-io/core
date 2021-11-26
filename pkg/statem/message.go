package statem

import (
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/mapper"
)

// PropertyMessage state property message.
type StateMessage struct {
	MessageBase

	StateID  string `json:"state_id"`
	Operator string `json:"operator"`
}

func NewPropertyMessage(id string, props map[string]constraint.Node) PropertyMessage {
	return PropertyMessage{
		StateID:    id,
		Operator:   "replace",
		Properties: props,
	}
}

// PropertyMessage state property message.
type PropertyMessage struct {
	MessageBase

	StateID    string                     `json:"state_id"`
	Operator   string                     `json:"operator"`
	Properties map[string]constraint.Node `json:"properties"`
}

type MapperMessage struct {
	MessageBase

	Operator string     `json:"operator"`
	Mapper   MapperDesc `json:"mapper"`
}

type TentacleMsg struct {
	MessageBase

	Operator string            `json:"operator"`
	StateID  string            `json:"state_id"`
	Items    []mapper.WatchKey `json:"items"`
}
