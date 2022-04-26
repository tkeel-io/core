package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	xerrors "github.com/tkeel-io/core/pkg/errors"
)

var repoIns IRepository
var testReady bool

//func TestNewExpression(t *testing.T) {
//	expr := NewExpression("admin", "device123", "expression001", "temp", "device234.temp", "")
//	assert.Equal(t, encodeKey("device123", "admin", "temp"), expr.ID)
//	t.Log("expressionID: ", expr.ID)
//}

func TestPutExpression(t *testing.T) {
	tests := []struct {
		name string
		expr *Expression
	}{
		{
			name: "test1",
			expr: NewExpression("admin", "device123", "expr1", "temp", "device002.temp", ""),
		},
		{
			name: "test2",
			expr: NewExpression("admin", "device123", "expr2", "cpu0", "device002.metrics.cpus[0].value", ""),
		},
		{
			name: "test3",
			expr: NewExpression("admin", "device123", "expr3", "mems[1].value", "device002.metrics.mems[1].value", ""),
		},
	}

	for _, exprInfo := range tests {
		t.Run(exprInfo.name, func(t *testing.T) {
			err := repoIns.PutExpression(context.Background(), *exprInfo.expr)
			assert.Nil(t, err)
		})
	}
}

func TestGetExpression(t *testing.T) {
	_, err := repoIns.GetExpression(context.Background(), Expression{EntityID: "device123", Owner: "admin", Path: "temp"})
	assert.ErrorIs(t, err, xerrors.ErrResourceNotFound)
	// assert.Equal(t, "admin", expr.Owner)
	// assert.Equal(t, "device123", expr.EntityID)
	// assert.Equal(t, "temp", expr.Path)
}

func TestListExpression(t *testing.T) {
	exprs, _ := repoIns.ListExpression(context.Background(),
		repoIns.GetLastRevision(context.Background()),
		&ListExprReq{EntityID: "device123", Owner: "admin"})
	t.Log("expressions: ", exprs)
}

func TestDelExpression(t *testing.T) {
	err := repoIns.DelExpression(context.Background(), Expression{Owner: "admin", EntityID: "device123", Path: "temp"})
	assert.Nil(t, err)
}

func TestDeleteExpressions(t *testing.T) {
	err := repoIns.DelExprByEnity(context.Background(), Expression{Owner: "admin", EntityID: "device123"})
	assert.Nil(t, err)
}
