package mapper

const (
	TentacleTypeUndefined = "undefined"
	TentacleTypeEntity    = "entity"
	TentacleTypeMapper    = "mapper"

	WatchKeyDelimiter = "."
)

type Mapper interface {
	// ID returns mapper id.
	ID() string
	// String returns MQL text.
	String() string
	// TargetEntity returns target entity.
	TargetEntity() string
	// SourceEntities returns source entities.
	SourceEntities() []string
	// Tentacles returns tentacles.
	Tentacles() []Tentacler
	// Copy duplicate a mapper.
	Copy() Mapper
	// Exec excute input returns output.
	Exec(map[string]interface{}) (map[string]interface{}, error)
}

type TentacleType = string

type Tentacler interface {
	// Type returns tentacle type.
	Type() TentacleType
	// TargetID returns target id.
	TargetID() string
	// Items returns watch keys(watchKey=entityId#propertyKey).
	Items() []WatchKey
	// Copy duplicate a mapper.
	Copy() Tentacler
	// IsRemote return remote flag.
	IsRemote() bool
}

type WatchKey struct {
	EntityId    string //nolint
	PropertyKey string
}

func (wk *WatchKey) String() string {
	return wk.EntityId + WatchKeyDelimiter + wk.PropertyKey
}
