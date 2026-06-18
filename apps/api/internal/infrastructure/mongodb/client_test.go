package mongodb

import (
	"context"
	"testing"
)

func TestNewClient_EmptyURI_ReturnsError(t *testing.T) {
	_, err := NewClient(context.Background(), "")
	if err == nil {
		t.Fatal("NewClient() error = nil, want error for empty uri")
	}
}
