package mapper

import "log"

type tentacle struct {
	tp       TentacleType
	remote   bool
	targetID string
	items    []WatchKey // key=entityId#propertyKey
}

func NewTentacle(tp TentacleType, targetID string, items []WatchKey) Tentacler {
	return &tentacle{
		tp:       tp,
		items:    items,
		targetID: targetID,
	}
}

func NewRemoteTentacle(tp TentacleType, targetID string, items []WatchKey) Tentacler {
	return &tentacle{
		tp:       tp,
		items:    items,
		remote:   true,
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
func (t *tentacle) Items() []WatchKey {
	return t.items
}

func (t *tentacle) Copy() Tentacler {
	items := make([]WatchKey, len(t.items))
	for index, item := range t.items {
		items[index] = item
	}

	return &tentacle{
		tp:       t.tp,
		items:    items,
		targetID: t.targetID,
	}
}

func (t *tentacle) IsRemote() bool {
	return t.remote
}

func MergeTentacles(tentacles ...Tentacler) Tentacler {
	if len(tentacles) == 0 {
		return nil
	}

	tentacle0, ok := tentacles[0].(*tentacle)
	if !ok {
		log.Fatalln("not want struct")
	}
	itemMap := make(map[string]WatchKey)
	for _, tentacle := range tentacles {
		for _, item := range tentacle.Items() {
			itemMap[item.String()] = item
		}
	}

	index := -1
	items := make([]WatchKey, len(itemMap))
	for _, item := range itemMap {
		index++
		items[index] = item
	}

	return NewTentacle(tentacle0.tp, tentacle0.targetID, items)
}
