package mapper

type tentacle struct {
	tp       TentacleType
	targetId string
	items    []string //key=entityId#propertyKey
}

func NewTentacle(tp TentacleType, targetId string, items []string) Tentacler {
	return &tentacle{
		tp:       tp,
		items:    items,
		targetId: targetId,
	}
}

// Type returns tentacle type.
func (t *tentacle) Type() TentacleType {
	return t.tp
}

// TargetId returns target id.
func (t *tentacle) TargetId() string {
	return t.targetId
}

// Items returns watch keys(watchKey=entityId#propertyKey).
func (t *tentacle) Items() []string {
	return t.items
}

func (t *tentacle) Copy() Tentacler {

	items := make([]string, len(t.items))
	for index, item := range t.items {
		items[index] = item
	}

	return &tentacle{
		tp:       t.tp,
		items:    items,
		targetId: t.targetId,
	}
}
