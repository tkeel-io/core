package mapper

type tentacle struct{}

func NewTentacle() *tentacle {
	return &tentacle{}
}

// Type returns tentacle type.
func (t *tentacle) Type() TentacleType {
	panic("implement me.")
}

// TargetId returns target id.
func (t *tentacle) TargetId() string {
	panic("implement me.")
}

// Items returns watch keys(watchKey=entityId#propertyKey).
func (t *tentacle) Items() []string {
	panic("implement me.")
}
