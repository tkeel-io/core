package constraint

// define里面存储的是关于property属性的定义.
type Define struct {
	Items []Itemer
}

func (d Define) Extract() map[string]interface{} {
	return nil
}

type Itemer interface {
	Check([]byte) bool
}

type item struct{}

func (mi item) CheckValue(value []byte) bool {
	return true
}

// MaxItem is a constraint for number.
type MaxItem struct {
	Max float64
}

func (mi MaxItem) CheckValue(value []byte) bool {
	return true
}

// MinItem is a constraint for number.
type MinItem struct {
	Min float64
}

func (mi MinItem) CheckValue(value []byte) bool {
	return true
}

// TypeItem for all types.
type TypeItem struct {
	Type string
}

func (pc TypeItem) CheckValue(value []byte) bool {
	return true
}

// UnitItem for basic types.
type UnitItem struct {
	item
	Unit   string
	UnitZH string
}

// LegthItem for array types.
type LegthItem struct {
	Length int
}

func (pc LegthItem) CheckValue(value []byte) bool {
	return true
}
