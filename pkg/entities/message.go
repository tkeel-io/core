package entities

type EntityMsg struct {
	SourceId string                 `json:"source_id"`
	Values   map[string]interface{} `json:"values"`
}

type TentacleMsg struct {
	TargetId string   `json:"target"`
	Items    []string `json:"items"`
}

func (em *EntityMsg) Message()   {}
func (em *TentacleMsg) Message() {}
