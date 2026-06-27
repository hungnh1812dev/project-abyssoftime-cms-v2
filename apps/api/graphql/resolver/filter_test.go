package resolver

import (
	"testing"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type testFilter struct {
	Title    *testStringFilter `json:"title"`
	Active   *testBoolFilter   `json:"active"`
	Position *testStringFilter `json:"position"`
	And      []*testFilter     `json:"and"`
	Or       []*testFilter     `json:"or"`
	Not      *testFilter       `json:"not"`
}

type testStringFilter struct {
	Eq    *string  `json:"eq"`
	Ne    *string  `json:"ne"`
	In    []string `json:"in"`
	NotIn []string `json:"notIn"`
}

type testBoolFilter struct {
	Eq *bool `json:"eq"`
	Ne *bool `json:"ne"`
}

func strPtr(val string) *string { return &val }
func boolPtr(val bool) *bool   { return &val }

func TestConvertFilterStructs_Nil(test *testing.T) {
	result := convertFilterStructs[testFilter](nil)
	if result != nil {
		test.Errorf("expected nil, got %v", result)
	}
}

func TestConvertFilterStructs_Empty(test *testing.T) {
	result := convertFilterStructs[testFilter]([]*testFilter{})
	if result != nil {
		test.Errorf("expected nil, got %v", result)
	}
}

func TestConvertFilterStructs_SingleFieldEq(test *testing.T) {
	filters := []*testFilter{
		{Title: &testStringFilter{Eq: strPtr("Hello")}},
	}
	result := convertFilterStructs(filters)
	if len(result) != 1 {
		test.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].Field == nil {
		test.Fatal("Field is nil")
	}
	if result[0].Field.Field != "title" {
		test.Errorf("Field = %q, want title", result[0].Field.Field)
	}
	if result[0].Field.Operator != entity.FilterOpEq {
		test.Errorf("Operator = %q, want eq", result[0].Field.Operator)
	}
	if result[0].Field.Value != "Hello" {
		test.Errorf("Value = %v, want Hello", result[0].Field.Value)
	}
}

func TestConvertFilterStructs_MultipleFilters(test *testing.T) {
	filters := []*testFilter{
		{Title: &testStringFilter{Eq: strPtr("A")}},
		{Position: &testStringFilter{Ne: strPtr("B")}},
	}
	result := convertFilterStructs(filters)
	if len(result) != 2 {
		test.Fatalf("len = %d, want 2", len(result))
	}
}

func TestConvertFilterStructs_OrCombinator(test *testing.T) {
	filters := []*testFilter{
		{Or: []*testFilter{
			{Position: &testStringFilter{Eq: strPtr("C")}},
			{Position: &testStringFilter{Eq: strPtr("D")}},
		}},
	}
	result := convertFilterStructs(filters)
	if len(result) != 1 {
		test.Fatalf("len = %d, want 1", len(result))
	}
	if len(result[0].Or) != 2 {
		test.Fatalf("Or len = %d, want 2", len(result[0].Or))
	}
	if result[0].Or[0].Field.Field != "position" {
		test.Errorf("Or[0].Field = %q, want position", result[0].Or[0].Field.Field)
	}
}

func TestConvertFilterStructs_NotCombinator(test *testing.T) {
	filters := []*testFilter{
		{Not: &testFilter{Title: &testStringFilter{Eq: strPtr("Hidden")}}},
	}
	result := convertFilterStructs(filters)
	if len(result) != 1 {
		test.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].Not == nil {
		test.Fatal("Not is nil")
	}
	if result[0].Not.Field == nil || result[0].Not.Field.Field != "title" {
		test.Errorf("Not.Field = %v, want title", result[0].Not)
	}
}

func TestConvertFilterStructs_NestedAndInsideOr(test *testing.T) {
	filters := []*testFilter{
		{Or: []*testFilter{
			{And: []*testFilter{
				{Title: &testStringFilter{Eq: strPtr("A")}},
				{Active: &testBoolFilter{Eq: boolPtr(true)}},
			}},
			{Position: &testStringFilter{Eq: strPtr("B")}},
		}},
	}
	result := convertFilterStructs(filters)
	if len(result) != 1 {
		test.Fatalf("len = %d, want 1", len(result))
	}
	if len(result[0].Or) != 2 {
		test.Fatalf("Or len = %d, want 2", len(result[0].Or))
	}
	if len(result[0].Or[0].And) != 2 {
		test.Errorf("Or[0].And len = %d, want 2", len(result[0].Or[0].And))
	}
}

func TestConvertFilterStructs_InOperator(test *testing.T) {
	filters := []*testFilter{
		{Title: &testStringFilter{In: []string{"a", "b", "c"}}},
	}
	result := convertFilterStructs(filters)
	if len(result) != 1 {
		test.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].Field.Operator != entity.FilterOpIn {
		test.Errorf("Operator = %q, want in", result[0].Field.Operator)
	}
	arr, isSlice := result[0].Field.Value.([]interface{})
	if !isSlice || len(arr) != 3 {
		test.Errorf("Value = %v, want [a b c]", result[0].Field.Value)
	}
}

func TestConvertFilterStructs_NotInOperator(test *testing.T) {
	filters := []*testFilter{
		{Title: &testStringFilter{NotIn: []string{"archived"}}},
	}
	result := convertFilterStructs(filters)
	if len(result) != 1 {
		test.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].Field.Operator != entity.FilterOpNotIn {
		test.Errorf("Operator = %q, want notIn", result[0].Field.Operator)
	}
}

func TestConvertFilterStructs_BooleanFilter(test *testing.T) {
	filters := []*testFilter{
		{Active: &testBoolFilter{Eq: boolPtr(true)}},
	}
	result := convertFilterStructs(filters)
	if len(result) != 1 {
		test.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].Field.Field != "active" {
		test.Errorf("Field = %q, want active", result[0].Field.Field)
	}
	if result[0].Field.Value != true {
		test.Errorf("Value = %v, want true", result[0].Field.Value)
	}
}

func TestPascalToCamelCase(test *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"DocumentID", "documentID"},
		{"Title", "title"},
		{"CreatedAt", "createdAt"},
		{"IsMain", "isMain"},
		{"URL", "url"},
	}
	for _, testCase := range tests {
		got := pascalToCamelCase(testCase.input)
		if got != testCase.want {
			test.Errorf("pascalToCamelCase(%q) = %q, want %q", testCase.input, got, testCase.want)
		}
	}
}

type testSortOrder string

const (
	testSortASC  testSortOrder = "ASC"
	testSortDESC testSortOrder = "DESC"
)

type testOrderBy struct {
	Title     *testSortOrder
	CreatedAt *testSortOrder
}

func TestExtractOrderBy_Nil(test *testing.T) {
	field, dir := extractOrderBy(nil)
	if field != "createdAt" || dir != -1 {
		test.Errorf("got (%q, %d), want (createdAt, -1)", field, dir)
	}
}

func TestExtractOrderBy_ASC(test *testing.T) {
	asc := testSortASC
	orderBy := &testOrderBy{Title: &asc}
	field, dir := extractOrderBy(orderBy)
	if field != "title" || dir != 1 {
		test.Errorf("got (%q, %d), want (title, 1)", field, dir)
	}
}

func TestExtractOrderBy_DESC(test *testing.T) {
	desc := testSortDESC
	orderBy := &testOrderBy{CreatedAt: &desc}
	field, dir := extractOrderBy(orderBy)
	if field != "createdAt" || dir != -1 {
		test.Errorf("got (%q, %d), want (createdAt, -1)", field, dir)
	}
}

func TestExtractOrderBy_FirstNonNilWins(test *testing.T) {
	asc := testSortASC
	desc := testSortDESC
	orderBy := &testOrderBy{Title: &asc, CreatedAt: &desc}
	field, _ := extractOrderBy(orderBy)
	if field != "title" {
		test.Errorf("field = %q, want title (first non-nil)", field)
	}
}
