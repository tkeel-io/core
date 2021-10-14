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

func MergeTentacles(tentacles ...Tentacler) Tentacler {

	if len(tentacles) == 0 {
		return nil
	}

	tentacle0 := tentacles[0].(*tentacle)

	itemMap := make(map[string]struct{})
	for _, tentacle := range tentacles {
		for _, item := range tentacle.Items() {
			itemMap[item] = struct{}{}
		}
	}

	index := -1
	items := make([]string, len(itemMap))
	for item, _ := range itemMap {
		index += 1
		items[index] = item
	}

	return NewTentacle(tentacle0.tp, tentacle0.targetId, items)
}
