package path

import (
	"strings"
)

const (
	Separator    = "."
	WildcardOne  = "+"
	WildcardSome = "*"
)

type Node interface {
	ID() string
	String() string
}

func New() *Tree {
	return &Tree{
		Separator:    Separator,
		WildcardOne:  WildcardOne,
		WildcardSome: WildcardSome,
		root:         newNode(),
	}
}

func fmtPath(path string) string {
	return strings.ReplaceAll(path, "[", ".[")
}

func FmtWatchKey(eid, propertyKey string) string {
	return eid + "." + propertyKey
}

func MergePath(subPath, changePath string) string {
	subPath = strings.TrimRight(subPath, Separator+WildcardSome)
	if strings.Contains(subPath, WildcardOne) {
		subPath = fmtPath(subPath)
		subSegs := strings.Split(subPath, Separator)
		changeSegs := strings.Split(changePath, Separator)
		if len(subSegs) > len(changeSegs) {
			return changePath
		}

		for index := range subSegs {
			if subSegs[index] == WildcardOne ||
				subSegs[index] == WildcardSome {
				subSegs[index] = changeSegs[index]
			}
		}
		return strings.Join(subSegs, Separator)
	}
	return subPath
}
