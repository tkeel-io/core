package search

import (
	"strings"
	"testing"
)

func TestStrings(t *testing.T) {
	t.Log(strings.ContainsAny("sasssss.xx[]", ".["))
}
