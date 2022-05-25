package expression

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/tdtl"
)

func Test_Expression(t *testing.T) {
	exprIns, err := NewExpr("device234.properties.temp2", nil)
	assert.Nil(t, err)
	t.Log(exprIns.Sources())
	assert.Equal(t, map[string][]string{"device234": {"device234.properties.temp2"}}, exprIns.Sources())
	res, err := exprIns.Eval(context.Background(),
		map[string]tdtl.Node{"device234.properties.temp2": tdtl.NewInt64(89)})
	assert.Nil(t, err)
	t.Log(res.String())
	assert.Equal(t, tdtl.IntNode(89), res)
	res, err = exprIns.Eval(context.Background(),
		map[string]tdtl.Node{"device234.properties.temp2.abc": tdtl.NewInt64(89)})
	assert.Nil(t, err)
	t.Log(res.String())
	assert.Equal(t, tdtl.UNDEFINED_RESULT, res)
}

func Test_Validate(t *testing.T) {
	type args struct {
		expr repository.Expression
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"1", args{
			expr: repository.Expression{Expression: "a.b.c"},
		}, false},
		{"2", args{
			expr: repository.Expression{Expression: "*.a"},
		}, true},
		{"3", args{
			expr: repository.Expression{Expression: "a.*"},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Validate(tt.args.expr); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
