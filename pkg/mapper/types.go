package mapper

/*

MQL:
	select
		entity1.property1 as property1,
		entity2.property2 as property2,
		entity1.property3 + entity2.property3 as property3,
	from
		entity3;

Parse:
	TargetEntity() returns:
		entity3
	SourceEntities() returns:
		entity1, entity2

Input:
	```{
		"entity3": {
			"property1": "",
			"property2": "",
			"property3": ""
		}
		"entity1": {
			"property1": "",
			"property2": "",
			"property3": ""
		}
		"entity2": {
			"property1": "",
			"property2": "",
			"property3": ""
		}
	}```
*/

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
}
