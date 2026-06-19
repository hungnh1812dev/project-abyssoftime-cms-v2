package grpcclient

import (
	"testing"
)

func TestManager_GetConnection_NotFound(t *testing.T) {
	m := NewManager()
	_, err := m.GetConnection("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown service")
	}
}

func TestManager_Close_Empty(t *testing.T) {
	m := NewManager()
	if err := m.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestParseServices(t *testing.T) {
	tests := []struct {
		raw  string
		want int
	}{
		{"", 0},
		{"search=localhost:9091", 1},
		{"search=localhost:9091,notify=localhost:9092", 2},
	}
	for _, tt := range tests {
		got := ParseServices(tt.raw)
		if len(got) != tt.want {
			t.Errorf("ParseServices(%q) = %d entries, want %d", tt.raw, len(got), tt.want)
		}
	}
}
