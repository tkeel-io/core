package expression

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/tdtl"
)

func Test_Expression(t *testing.T) {
	exprIns, err := NewExpr("a.device123", nil)
	assert.Nil(t, err)
	t.Log(exprIns.Entities())
	res, err := exprIns.Eval(context.Background(), map[string]tdtl.Node{"a.device123": tdtl.NewInt64(20)})
	assert.Nil(t, err)
	t.Log(res.String())
}
