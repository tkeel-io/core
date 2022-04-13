package dao

import (
	"crypto/md5"
)

const (
	// owner/entityID/id
	fmtExprString = "/core/v1/expression/%s/%s/%s"
)


type Expression struct {
	ID string
	// expression owner.
	Owner string
	// entity id.
	EntityID string
	// expression.
	Expression string
}


func NewExpression(owner, entityID, expr string) *Expression {
	return &Expression {
		ID: util.UUID(),
		Owner: owner,
		EntityID: entityID,
		Expression: expr,
	}
}



func (e *Expression) Key() string {
	return fmt.Sprintf(fmtExprString, e.Owner, e.EntityID, e.Expression)
}
 


func (d* dao) PutExpression(ctx context.Context, expr Expression) error {
	var err error
	var bytes []byte
	if bytes, err = json.Marshal(expr); nil == err {
		_, err = d.etcdEndpoint.Put(ctx, m.Key(), string(bytes))
	}
	return errors.Wrap(err, "put expression")
}


func (d* dao) GetExpression(ctx context.Context, expr Expression) (Expression, error) {
	
}
