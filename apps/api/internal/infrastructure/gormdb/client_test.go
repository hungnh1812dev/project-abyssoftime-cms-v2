package gormdb

import (
	"testing"
)

func TestNewClient_InvalidDSN(t *testing.T) {
	_, err := NewClient("postgres", "host=invalid port=0 dbname=nonexistent sslmode=disable connect_timeout=1")
	if err == nil {
		t.Fatal("expected error for invalid DSN, got nil")
	}
}

func TestNewClient_UnsupportedDriver(t *testing.T) {
	_, err := NewClient("oracle", "fake-dsn")
	if err == nil {
		t.Fatal("expected error for unsupported driver, got nil")
	}
}

func TestNewClient_SQLite(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("NewClient(sqlite) error: %v", err)
	}
	if db == nil {
		t.Fatal("expected non-nil db")
	}
}

// AutoMigrate is tested after C-2 adds GORM struct tags to entities.
