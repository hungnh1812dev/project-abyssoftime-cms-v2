package dynamic

import (
	"testing"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

func TestParseFilters_Nil(t *testing.T) {
	result := parseFilters(nil)
	if result != nil {
		t.Errorf("parseFilters(nil) = %v, want nil", result)
	}
}

func TestParseFilters_EmptySlice(t *testing.T) {
	result := parseFilters([]any{})
	if result != nil {
		t.Errorf("parseFilters([]) = %v, want nil", result)
	}
}

func TestParseFilters_SingleFieldEq(t *testing.T) {
	input := []any{
		map[string]any{
			"documentId": map[string]any{"eq": "abc-123"},
		},
	}
	result := parseFilters(input)
	if len(result) != 1 {
		t.Fatalf("parseFilters() len = %d, want 1", len(result))
	}
	if result[0].Field == nil {
		t.Fatal("parseFilters() result[0].Field is nil")
	}
	if result[0].Field.Field != "documentId" {
		t.Errorf("Field = %q, want documentId", result[0].Field.Field)
	}
	if result[0].Field.Operator != entity.FilterOpEq {
		t.Errorf("Operator = %q, want eq", result[0].Field.Operator)
	}
	if result[0].Field.Value != "abc-123" {
		t.Errorf("Value = %v, want abc-123", result[0].Field.Value)
	}
}

func TestParseFilters_MultipleFilters_ImplicitAnd(t *testing.T) {
	input := []any{
		map[string]any{"documentId": map[string]any{"eq": "A"}},
		map[string]any{"title": map[string]any{"ne": "Draft"}},
	}
	result := parseFilters(input)
	if len(result) != 2 {
		t.Fatalf("parseFilters() len = %d, want 2", len(result))
	}
	foundDocID := false
	foundTitle := false
	for _, node := range result {
		if node.Field != nil && node.Field.Field == "documentId" {
			foundDocID = true
		}
		if node.Field != nil && node.Field.Field == "title" {
			foundTitle = true
		}
	}
	if !foundDocID || !foundTitle {
		t.Errorf("expected both documentId and title filters, got %+v", result)
	}
}

func TestParseFilters_OrCombinator(t *testing.T) {
	input := []any{
		map[string]any{
			"or": []any{
				map[string]any{"position": map[string]any{"eq": "C"}},
				map[string]any{"position": map[string]any{"eq": "D"}},
			},
		},
	}
	result := parseFilters(input)
	if len(result) != 1 {
		t.Fatalf("parseFilters() len = %d, want 1", len(result))
	}
	if len(result[0].Or) != 2 {
		t.Fatalf("Or len = %d, want 2", len(result[0].Or))
	}
	if result[0].Or[0].Field.Field != "position" {
		t.Errorf("Or[0].Field = %q, want position", result[0].Or[0].Field.Field)
	}
}

func TestParseFilters_NotCombinator(t *testing.T) {
	input := []any{
		map[string]any{
			"not": map[string]any{
				"title": map[string]any{"eq": "Hidden"},
			},
		},
	}
	result := parseFilters(input)
	if len(result) != 1 {
		t.Fatalf("parseFilters() len = %d, want 1", len(result))
	}
	if result[0].Not == nil {
		t.Fatal("Not is nil")
	}
	if result[0].Not.Field == nil || result[0].Not.Field.Field != "title" {
		t.Errorf("Not.Field = %v, want title", result[0].Not)
	}
}

func TestParseFilters_NestedAndInsideOr(t *testing.T) {
	input := []any{
		map[string]any{
			"or": []any{
				map[string]any{
					"and": []any{
						map[string]any{"title": map[string]any{"eq": "A"}},
						map[string]any{"active": map[string]any{"eq": true}},
					},
				},
				map[string]any{"position": map[string]any{"eq": "B"}},
			},
		},
	}
	result := parseFilters(input)
	if len(result) != 1 {
		t.Fatalf("parseFilters() len = %d, want 1", len(result))
	}
	if len(result[0].Or) != 2 {
		t.Fatalf("Or len = %d, want 2", len(result[0].Or))
	}
	if len(result[0].Or[0].And) != 2 {
		t.Errorf("Or[0].And len = %d, want 2", len(result[0].Or[0].And))
	}
}

func TestParseFilters_InOperator(t *testing.T) {
	input := []any{
		map[string]any{
			"documentId": map[string]any{"in": []any{"a", "b", "c"}},
		},
	}
	result := parseFilters(input)
	if len(result) != 1 {
		t.Fatalf("parseFilters() len = %d, want 1", len(result))
	}
	if result[0].Field.Operator != entity.FilterOpIn {
		t.Errorf("Operator = %q, want in", result[0].Field.Operator)
	}
}

func TestParseFilters_NotInOperator(t *testing.T) {
	input := []any{
		map[string]any{
			"status": map[string]any{"notIn": []any{"archived"}},
		},
	}
	result := parseFilters(input)
	if len(result) != 1 {
		t.Fatalf("parseFilters() len = %d, want 1", len(result))
	}
	if result[0].Field.Operator != entity.FilterOpNotIn {
		t.Errorf("Operator = %q, want notIn", result[0].Field.Operator)
	}
}

func TestParseFilters_MixedFieldAndLogical(t *testing.T) {
	input := []any{
		map[string]any{
			"title": map[string]any{"eq": "Hello"},
			"or": []any{
				map[string]any{"active": map[string]any{"eq": true}},
			},
		},
	}
	result := parseFilters(input)
	if len(result) != 2 {
		t.Fatalf("parseFilters() len = %d, want 2 (one field + one or)", len(result))
	}
}
