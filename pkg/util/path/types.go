package path

import (
	"strings"
)

type Node interface {
	ID() string
	String() string
}

func New() *Tree {
	return &Tree{
		Separator:    ".",
		WildcardOne:  "+",
		WildcardSome: "*",

		root: newNode(),
	}
}

func fmtPath(path string) string {
	return strings.ReplaceAll(path, "[", ".[")
}

func FmtWatchKey(eid, propertyKey string) string {
	return eid + "." + propertyKey
}
