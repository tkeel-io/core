package mapper

import "fmt"

const (
	TentacleTypeUndefine = "undefine"
	TentacleTypeEntity   = "entity"
	TentacleTypeMapper   = "mapper"
)

type Mapper interface {
	// Id returns mapper id.
	Id() string
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
	// TargetId returns target id.
	TargetId() string
	// Items returns watch keys(watchKey=entityId#propertyKey).
	Items() []string
	// Copy duplicate a mapper.
	Copy() Tentacler
}

func GenTentacleKey(entityId, propertyKey string) string {
	return fmt.Sprintf("%s#%s", entityId, propertyKey)
}
