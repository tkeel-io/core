package util

import "fmt"

const (
	fmtMapperString = "core.%s.mapper.%s.%s" // core.type.mapper.entityID.name.
)

func FormatMapper(typ, id, name string) string {
	return fmt.Sprintf(fmtMapperString, typ, id, name)
}
