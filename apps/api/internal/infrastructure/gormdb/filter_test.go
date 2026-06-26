package gormdb

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	return database
}

func TestApplyFilters_Nil(t *testing.T) {
	database := openTestDB(t)
	query := database.Table("test")
	result := applyFilters(query, nil)
	if result != query {
		t.Error("applyFilters(nil) should return same query")
	}
}

func TestApplyFilters_Empty(t *testing.T) {
	database := openTestDB(t)
	query := database.Table("test")
	result := applyFilters(query, []entity.FilterNode{})
	if result != query {
		t.Error("applyFilters([]) should return same query")
	}
}

func TestFilterFieldToColumn_SystemFields(t *testing.T) {
	tests := []struct {
		field string
		want  string
	}{
		{"documentId", "document_id"},
		{"createdAt", "created_at"},
		{"updatedAt", "updated_at"},
		{"publishedAt", "published_at"},
	}
	for _, test := range tests {
		col, ok := filterFieldToColumn(test.field)
		if !ok {
			t.Errorf("filterFieldToColumn(%q) not ok", test.field)
		}
		if col != test.want {
			t.Errorf("filterFieldToColumn(%q) = %q, want %q", test.field, col, test.want)
		}
	}
}

func TestFilterFieldToColumn_ContentField(t *testing.T) {
	col, ok := filterFieldToColumn("title")
	if !ok {
		t.Error("filterFieldToColumn(title) not ok")
	}
	if col != "title" {
		t.Errorf("filterFieldToColumn(title) = %q, want title", col)
	}
}

func TestFilterFieldToColumn_InvalidChars(t *testing.T) {
	_, ok := filterFieldToColumn("title; DROP TABLE")
	if ok {
		t.Error("filterFieldToColumn should reject field with special chars")
	}
}

func TestFilterFieldToColumn_Empty(t *testing.T) {
	_, ok := filterFieldToColumn("")
	if ok {
		t.Error("filterFieldToColumn should reject empty field")
	}
}

func TestApplyFilters_SingleEq(t *testing.T) {
	database := openTestDB(t)
	database.Exec("CREATE TABLE test (document_id TEXT, title TEXT)")
	database.Exec("INSERT INTO test VALUES ('a', 'hello')")
	database.Exec("INSERT INTO test VALUES ('b', 'world')")

	filters := []entity.FilterNode{
		{Field: &entity.FieldFilter{Field: "documentId", Operator: entity.FilterOpEq, Value: "a"}},
	}

	var rows []map[string]any
	query := applyFilters(database.Table("test"), filters)
	query.Find(&rows)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
}

func TestApplyFilters_Ne(t *testing.T) {
	database := openTestDB(t)
	database.Exec("CREATE TABLE test (title TEXT)")
	database.Exec("INSERT INTO test VALUES ('hello')")
	database.Exec("INSERT INTO test VALUES ('world')")

	filters := []entity.FilterNode{
		{Field: &entity.FieldFilter{Field: "title", Operator: entity.FilterOpNe, Value: "hello"}},
	}

	var rows []map[string]any
	query := applyFilters(database.Table("test"), filters)
	query.Find(&rows)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
}

func TestApplyFilters_In(t *testing.T) {
	database := openTestDB(t)
	database.Exec("CREATE TABLE test (document_id TEXT)")
	database.Exec("INSERT INTO test VALUES ('a')")
	database.Exec("INSERT INTO test VALUES ('b')")
	database.Exec("INSERT INTO test VALUES ('c')")

	filters := []entity.FilterNode{
		{Field: &entity.FieldFilter{Field: "documentId", Operator: entity.FilterOpIn, Value: []string{"a", "c"}}},
	}

	var rows []map[string]any
	query := applyFilters(database.Table("test"), filters)
	query.Find(&rows)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
}

func TestApplyFilters_NotIn(t *testing.T) {
	database := openTestDB(t)
	database.Exec("CREATE TABLE test (document_id TEXT)")
	database.Exec("INSERT INTO test VALUES ('a')")
	database.Exec("INSERT INTO test VALUES ('b')")
	database.Exec("INSERT INTO test VALUES ('c')")

	filters := []entity.FilterNode{
		{Field: &entity.FieldFilter{Field: "documentId", Operator: entity.FilterOpNotIn, Value: []string{"a", "c"}}},
	}

	var rows []map[string]any
	query := applyFilters(database.Table("test"), filters)
	query.Find(&rows)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
}

func TestApplyFilters_OrCombinator(t *testing.T) {
	database := openTestDB(t)
	database.Exec("CREATE TABLE test (title TEXT)")
	database.Exec("INSERT INTO test VALUES ('hello')")
	database.Exec("INSERT INTO test VALUES ('world')")
	database.Exec("INSERT INTO test VALUES ('other')")

	filters := []entity.FilterNode{
		{Or: []entity.FilterNode{
			{Field: &entity.FieldFilter{Field: "title", Operator: entity.FilterOpEq, Value: "hello"}},
			{Field: &entity.FieldFilter{Field: "title", Operator: entity.FilterOpEq, Value: "world"}},
		}},
	}

	var rows []map[string]any
	query := applyFilters(database.Table("test"), filters)
	query.Find(&rows)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
}

func TestApplyFilters_AndCombinator(t *testing.T) {
	database := openTestDB(t)
	database.Exec("CREATE TABLE test (title TEXT, active INTEGER)")
	database.Exec("INSERT INTO test VALUES ('hello', 1)")
	database.Exec("INSERT INTO test VALUES ('hello', 0)")
	database.Exec("INSERT INTO test VALUES ('world', 1)")

	filters := []entity.FilterNode{
		{And: []entity.FilterNode{
			{Field: &entity.FieldFilter{Field: "title", Operator: entity.FilterOpEq, Value: "hello"}},
			{Field: &entity.FieldFilter{Field: "active", Operator: entity.FilterOpEq, Value: 1}},
		}},
	}

	var rows []map[string]any
	query := applyFilters(database.Table("test"), filters)
	query.Find(&rows)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
}

func TestApplyFilters_InvalidFieldSkipped(t *testing.T) {
	database := openTestDB(t)
	database.Exec("CREATE TABLE test (title TEXT)")
	database.Exec("INSERT INTO test VALUES ('hello')")

	filters := []entity.FilterNode{
		{Field: &entity.FieldFilter{Field: "bad;field", Operator: entity.FilterOpEq, Value: "x"}},
	}

	var rows []map[string]any
	query := applyFilters(database.Table("test"), filters)
	query.Find(&rows)
	if len(rows) != 1 {
		t.Fatalf("invalid field should be skipped, expected 1 row, got %d", len(rows))
	}
}
