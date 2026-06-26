package entity

type FilterOperator string

const (
	FilterOpEq    FilterOperator = "eq"
	FilterOpNe    FilterOperator = "ne"
	FilterOpIn    FilterOperator = "in"
	FilterOpNotIn FilterOperator = "notIn"
)

type FieldFilter struct {
	Field    string
	Operator FilterOperator
	Value    any
}

type FilterNode struct {
	Field *FieldFilter
	And   []FilterNode
	Or    []FilterNode
	Not   *FilterNode
}
