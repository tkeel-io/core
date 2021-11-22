package constraint

import "errors"

var (
	ErrEntityConfigInvalid = errors.New("invalid entity configurations")
)

type Constraint struct {
	Operator string
	JSONPath string
	Cond     Itemer
}
