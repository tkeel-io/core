package tql

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
func (t *tql) Exec(in map[string][]byte) (map[string][]byte, error) {
	return t.listener.GetComputeResults(in), nil
}
