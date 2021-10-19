package mapper

import "fmt"

const (
	TentacleTypeUndefined = "undefined"
	TentacleTypeEntity    = "entity"
	TentacleTypeMapper    = "mapper"
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
	Exec(map[string]map[string]interface{}) (map[string]map[string]interface{}, error)
}

type TentacleType = string

type Tentacler interface {
	// Type returns tentacle type.
	Type() TentacleType
	// TargetID returns target id.
	TargetID() string
	// Items returns watch keys(watchKey=entityId#propertyKey).
	Items() []string
	// Copy duplicate a mapper.
	Copy() Tentacler
}

func GenTentacleKey(entityID, propertyKey string) string {
	return fmt.Sprintf("%s#%s", entityID, propertyKey)
}
