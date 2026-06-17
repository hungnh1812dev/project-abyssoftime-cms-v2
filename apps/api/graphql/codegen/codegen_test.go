package codegen_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"project-abyssoftime-cms-v2/api/graphql/codegen"
)

func writeSchema(t *testing.T, dir, content string) string {
	t.Helper()
	p := filepath.Join(dir, "schema.graphqls")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func writeSentinel(t *testing.T, dir, hash string) string {
	t.Helper()
	p := filepath.Join(dir, ".graphql.hash")
	if err := os.WriteFile(p, []byte(hash), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestEnsureUpToDate_HashMatch(t *testing.T) {
	dir := t.TempDir()
	schemaContent := "type Query { hello: String }"
	schemaPath := writeSchema(t, dir, schemaContent)

	// Pre-compute correct hash by running EnsureUpToDate once with a recording runner
	sentinelPath := filepath.Join(dir, ".graphql.hash")
	called := 0
	if err := codegen.EnsureUpToDate(schemaPath, sentinelPath, func() error { called++; return nil }); err != nil {
		t.Fatalf("first run: %v", err)
	}
	if called != 1 {
		t.Fatalf("expected runner called once on first run, got %d", called)
	}

	// Second run — hash matches, runner must NOT be called
	called = 0
	if err := codegen.EnsureUpToDate(schemaPath, sentinelPath, func() error { called++; return nil }); err != nil {
		t.Fatalf("second run: %v", err)
	}
	if called != 0 {
		t.Errorf("runner should not be called when hash matches, got %d calls", called)
	}
}

func TestEnsureUpToDate_HashMismatch(t *testing.T) {
	dir := t.TempDir()
	schemaPath := writeSchema(t, dir, "type Query { hello: String }")
	sentinelPath := writeSentinel(t, dir, "stale-hash-value")

	called := 0
	if err := codegen.EnsureUpToDate(schemaPath, sentinelPath, func() error { called++; return nil }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called != 1 {
		t.Errorf("expected runner called once on hash mismatch, got %d", called)
	}

	// Sentinel must now hold the correct hash
	data, err := os.ReadFile(sentinelPath)
	if err != nil {
		t.Fatalf("sentinel not written: %v", err)
	}
	if string(data) == "stale-hash-value" {
		t.Error("sentinel still holds stale hash after successful run")
	}
}

func TestEnsureUpToDate_NoSentinel(t *testing.T) {
	dir := t.TempDir()
	schemaPath := writeSchema(t, dir, "type Query { hello: String }")
	sentinelPath := filepath.Join(dir, ".graphql.hash") // does not exist

	called := 0
	if err := codegen.EnsureUpToDate(schemaPath, sentinelPath, func() error { called++; return nil }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called != 1 {
		t.Errorf("expected runner called once when sentinel absent, got %d", called)
	}

	if _, err := os.Stat(sentinelPath); err != nil {
		t.Errorf("sentinel file not created: %v", err)
	}
}

func TestEnsureUpToDate_RunnerFailure(t *testing.T) {
	dir := t.TempDir()
	schemaPath := writeSchema(t, dir, "type Query { hello: String }")
	sentinelPath := filepath.Join(dir, ".graphql.hash")

	runErr := errors.New("codegen failed")
	err := codegen.EnsureUpToDate(schemaPath, sentinelPath, func() error { return runErr })
	if err == nil {
		t.Fatal("expected error when runner fails, got nil")
	}

	// Sentinel must NOT be written on failure
	if _, statErr := os.Stat(sentinelPath); statErr == nil {
		t.Error("sentinel must not be written when runner fails")
	}
}
