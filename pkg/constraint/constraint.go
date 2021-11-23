package constraint

import (
	"errors"
)

const (
	ConstraintOpSearchFlush     = "searchCB"
	ConstraintOpTimeSeriesFlush = "timeseriesCB"
	ConstraintOpTypeConvert     = "convert"
)

var (
	ErrEntityConfigInvalid = errors.New("invalid entity configurations")
)

type Constraint struct {
	Operator string
	JSONPath string
	Conds    []Itemer
}

func NewConstraintsFrom(cfg Config) []Constraint {
	return []Constraint{}
}

func ParseFlushableFrom(cfg Config) (tsSlice []string, searchSlice []string) {
	// current latyer.
	if !cfg.Enabled {
		return
	}

	return parseFlushableFrom("", cfg)
}

func parseFlushableFrom(prefix string, cfg Config) (tsSlice []string, searchSlice []string) {
	// current latyer.
	if !cfg.Enabled {
		return
	}
	if cfg.EnabledSearch {
		searchSlice = append(searchSlice, prefix+cfg.ID)
	}
	if cfg.EnabledTimeSeries {
		tsSlice = append(tsSlice, prefix+cfg.ID)
	}

	switch cfg.Type {
	case PropertyTypeArray:
		define := cfg.getArrayDefine()
		tss, ss := parseFlushableFrom(cfg.ID+".", define.ElemType)
		tsSlice, searchSlice =
			append(tsSlice, tss...), append(searchSlice, ss...)
	case PropertyTypeStruct:
		define := cfg.getStructDefine()
		for _, field := range define.Fields {
			tss, ss := parseFlushableFrom(cfg.ID+".", field)
			tsSlice, searchSlice =
				append(tsSlice, tss...), append(searchSlice, ss...)
		}
	default:
		// TODO: .
	}
	return tsSlice, searchSlice
}

func ExecData(val Node, cts []Constraint) (Node, error) {
	return val, nil
}
