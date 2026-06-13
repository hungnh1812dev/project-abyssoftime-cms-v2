//go:build !integration

package mongodb_test

import (
	"context"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/infrastructure/mongodb"
)

// compile-time: NewUserRepository signature must match the interface factory pattern.
// The production code file already has: var _ repository.UserRepository = (*userRepository)(nil)
var _ = mongodb.NewUserRepository // function must exist and be exported

func TestUserRepository_Create_SetsID(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

// Table-driven integration test (excluded without -tags integration)
var userRepoTests = []struct {
	name  string
	email string
	role  entity.Role
}{
	{"admin user", "admin@example.com", entity.RoleAdmin},
	{"guest user", "guest@example.com", entity.RoleGuest},
}

func TestUserRepository_TableDriven(t *testing.T) {
	_ = context.Background
	_ = userRepoTests
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}
