package mapper

import "log"

type tentacle struct {
	tp       TentacleType
	targetID string
	items    []string // key=entityId#propertyKey
}

func NewTentacle(tp TentacleType, targetID string, items []string) Tentacler {
	return &tentacle{
		tp:       tp,
		items:    items,
		targetID: targetID,
	}
}

// Type returns tentacle type.
func (t *tentacle) Type() TentacleType {
	return t.tp
}

// TargetID returns target id.
func (t *tentacle) TargetID() string {
	return t.targetID
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
		targetID: t.targetID,
	}
}

func MergeTentacles(tentacles ...Tentacler) Tentacler {
	if len(tentacles) == 0 {
		return nil
	}

	tentacle0, ok := tentacles[0].(*tentacle)
	if !ok {
		log.Fatalln("not want struct")
	}
	itemMap := make(map[string]struct{})
	for _, tentacle := range tentacles {
		for _, item := range tentacle.Items() {
			itemMap[item] = struct{}{}
		}
	}

	index := -1
	items := make([]string, len(itemMap))
	for item := range itemMap {
		index++
		items[index] = item
	}

	return NewTentacle(tentacle0.tp, tentacle0.targetID, items)
}
