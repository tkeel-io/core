package expression

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/tdtl"
)

func Test_Expression(t *testing.T) {
	exprIns, err := NewExpr("device234.properties.temp2", nil)
	assert.Nil(t, err)
	t.Log(exprIns.Entities())
	res, err := exprIns.Eval(context.Background(),
		map[string]tdtl.Node{"device234.properties.temp2": tdtl.NewInt64(89)})
	t.Log(map[string]tdtl.Node{"device234.properties.temp2": tdtl.NewInt64(89)})
	assert.Nil(t, err)
	t.Log(res.String())
}
