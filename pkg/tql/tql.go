package tql

import (
	"github.com/tkeel-io/core/pkg/constraint"
)

type tql struct {
	text     string
	config   TQLConfig
	listener Listener
}

func NewTQL(tqlString string) TQL {
	listener := Parse(tqlString)
	return &tql{
		text:     tqlString,
		listener: listener,
		config:   listener.GetParseConfigs(),
	}
}

// Target returns target entity.
func (t *tql) Target() string {
	return t.config.TargetEntity
}

// Entities returns source entities.
func (t *tql) Entities() []string {
	return t.config.SourceEntities
}

// Tentacles returns tentacles.
func (t *tql) Tentacles() []TentacleConfig {
	return t.config.Tentacles
}

// Exec execute MQL.
func (t *tql) Exec(in map[string]constraint.Node) (map[string]constraint.Node, error) {
	input := make(map[string][]byte)
	for key, val := range in {
		input[key] = []byte(val.String())
	}
	ret := t.listener.GetComputeResults(input)

	out := make(map[string]constraint.Node)
	for key, val := range ret {
		out[key] = constraint.RawNode(val)
	}

	return out, nil
}
