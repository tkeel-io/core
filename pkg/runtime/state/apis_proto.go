package state

import "github.com/tkeel-io/core/pkg/constraint"

type PatchData struct {
	Path     string                   `json:"path"`
	Operator constraint.PatchOperator `json:"operator"`
	Value    interface{}              `json:"value"`
}
