package constraint

import "errors"

const (
	ConstraintOpSearchFlush     = "flush-search"
	ConstraintOpTimeSeriesFlush = "flush-timeseries"
)

var (
	ErrEntityConfigInvalid = errors.New("invalid entity configurations")
)

type Constraint struct {
	Operator string
	JSONPath string
	Cond     Itemer
}

func ExecData(val Node, cts []Constraint) (Node, error) {
	return val, nil
}
