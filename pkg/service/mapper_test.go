package service

import (
	"testing"

	"github.com/tkeel-io/tdtl"
)

func TestRaw(t *testing.T) {
	t.Log(string(tdtl.NewString("hahah").Raw()))
}
