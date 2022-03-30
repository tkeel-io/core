package util

import (
	"testing"
)

func TestResolveAddr(t *testing.T) {
	addr := ResolveAddr()
	t.Log(addr)
}
