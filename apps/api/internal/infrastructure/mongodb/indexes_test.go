//go:build !integration

package mongodb_test

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"

	"project-abyssoftime-cms-v2/api/internal/infrastructure/mongodb"
)

// compile-time: EnsureIndexes must exist with the expected signature.
var _ func(context.Context, *mongo.Database) error = mongodb.EnsureIndexes

func TestEnsureIndexes_Signature(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}
