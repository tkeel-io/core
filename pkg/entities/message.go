package entities

type EntityMsg struct {
	SourceID string                 `json:"source_id"`
	Values   map[string]interface{} `json:"values"`
}

type TentacleMsg struct {
	TargetID string   `json:"target"` //nolint
	Items    []string `json:"items"`
}

func (em *EntityMsg) Message()   {}
func (em *TentacleMsg) Message() {}
