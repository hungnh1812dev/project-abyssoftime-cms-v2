package role

import (
	"context"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository/mock"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

func seedRepo() *mock.RoleRepository {
	roles := map[string]*entity.RoleEntity{}
	return &mock.RoleRepository{
		CreateFn: func(_ context.Context, r *entity.RoleEntity) error {
			if _, exists := roles[r.Slug]; exists {
				return pkgerrors.ErrConflict
			}
			roles[r.Slug] = r
			return nil
		},
		FindByIDFn: func(_ context.Context, docID string) (*entity.RoleEntity, error) {
			for _, r := range roles {
				if r.DocumentID == docID {
					return r, nil
				}
			}
			return nil, pkgerrors.ErrNotFound
		},
		FindBySlugFn: func(_ context.Context, slug string) (*entity.RoleEntity, error) {
			if r, ok := roles[slug]; ok {
				return r, nil
			}
			return nil, pkgerrors.ErrNotFound
		},
		FindAllFn: func(_ context.Context) ([]*entity.RoleEntity, error) {
			out := make([]*entity.RoleEntity, 0, len(roles))
			for _, r := range roles {
				out = append(out, r)
			}
			return out, nil
		},
		UpdateFn: func(_ context.Context, r *entity.RoleEntity) error {
			roles[r.Slug] = r
			return nil
		},
		DeleteFn: func(_ context.Context, docID string) error {
			for slug, r := range roles {
				if r.DocumentID == docID {
					delete(roles, slug)
					return nil
				}
			}
			return pkgerrors.ErrNotFound
		},
		HasAnyFn: func(_ context.Context) (bool, error) {
			return len(roles) > 0, nil
		},
	}
}

func emptyUserRepo() *mock.UserRepository {
	return &mock.UserRepository{
		FindAllFn: func(_ context.Context, _, _ int) ([]*entity.User, int64, error) {
			return nil, 0, nil
		},
	}
}

func TestSeedDefaultRoles(t *testing.T) {
	repo := seedRepo()
	userRepo := emptyUserRepo()
	uc := New(repo, userRepo)

	if err := uc.SeedDefaults(context.Background()); err != nil {
		t.Fatalf("SeedDefaults: %v", err)
	}

	roles, _ := repo.FindAll(context.Background())
	if len(roles) != 4 {
		t.Errorf("seeded %d roles, want 4", len(roles))
	}

	// Idempotent: calling again should not fail
	if err := uc.SeedDefaults(context.Background()); err != nil {
		t.Fatalf("SeedDefaults (second call): %v", err)
	}

	roles, _ = repo.FindAll(context.Background())
	if len(roles) != 4 {
		t.Errorf("after second seed: %d roles, want 4", len(roles))
	}
}

func TestCreate_Valid(t *testing.T) {
	repo := seedRepo()
	userRepo := emptyUserRepo()
	uc := New(repo, userRepo)
	_ = uc.SeedDefaults(context.Background())

	role, err := uc.Create(context.Background(), CreateRoleInput{
		Name:        "Content Manager",
		Slug:        "content-manager",
		Permissions: []string{"content:read", "content:create"},
		Level:       50,
	}, 100) // caller level 100 (super_admin)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if role.Name != "Content Manager" {
		t.Errorf("Name = %q, want %q", role.Name, "Content Manager")
	}
	if role.DocumentID == "" {
		t.Error("expected DocumentID to be generated")
	}
}

func TestCreate_DuplicateSlug(t *testing.T) {
	repo := seedRepo()
	userRepo := emptyUserRepo()
	uc := New(repo, userRepo)
	_ = uc.SeedDefaults(context.Background())

	_, err := uc.Create(context.Background(), CreateRoleInput{
		Name: "Editor Clone", Slug: "editor",
		Permissions: []string{"content:read"}, Level: 50,
	}, 100)
	if !pkgerrors.Is(err, pkgerrors.ErrConflict) {
		t.Errorf("err = %v, want ErrConflict", err)
	}
}

func TestCreate_LevelTooHigh(t *testing.T) {
	repo := seedRepo()
	userRepo := emptyUserRepo()
	uc := New(repo, userRepo)

	_, err := uc.Create(context.Background(), CreateRoleInput{
		Name: "Overreach", Slug: "overreach",
		Permissions: []string{"content:read"}, Level: 80,
	}, 60) // caller level 60, trying to create level 80
	if !pkgerrors.Is(err, pkgerrors.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestCreate_InvalidPermission(t *testing.T) {
	repo := seedRepo()
	userRepo := emptyUserRepo()
	uc := New(repo, userRepo)

	_, err := uc.Create(context.Background(), CreateRoleInput{
		Name: "Bad", Slug: "bad",
		Permissions: []string{"content:read", "invalid:perm"}, Level: 50,
	}, 100)
	if !pkgerrors.Is(err, pkgerrors.ErrValidation) {
		t.Errorf("err = %v, want ErrValidation", err)
	}
}

func TestUpdate_DefaultRole_OnlyPermissions(t *testing.T) {
	repo := seedRepo()
	userRepo := emptyUserRepo()
	uc := New(repo, userRepo)
	_ = uc.SeedDefaults(context.Background())

	editorRole, _ := repo.FindBySlug(context.Background(), "editor")

	// Updating permissions should succeed
	updated, err := uc.Update(context.Background(), editorRole.DocumentID, UpdateRoleInput{
		Permissions: &[]string{"content:read", "content:create"},
	}, 100)
	if err != nil {
		t.Fatalf("Update permissions: %v", err)
	}
	if len(updated.Permissions) != 2 {
		t.Errorf("len(Permissions) = %d, want 2", len(updated.Permissions))
	}

	// Updating level on default role should fail
	newLevel := 70
	_, err = uc.Update(context.Background(), editorRole.DocumentID, UpdateRoleInput{
		Level: &newLevel,
	}, 100)
	if !pkgerrors.Is(err, pkgerrors.ErrValidation) {
		t.Errorf("err = %v, want ErrValidation for level change on default role", err)
	}
}

func TestUpdate_CustomRole_AllFields(t *testing.T) {
	repo := seedRepo()
	userRepo := emptyUserRepo()
	uc := New(repo, userRepo)
	_ = uc.SeedDefaults(context.Background())

	created, _ := uc.Create(context.Background(), CreateRoleInput{
		Name: "Custom", Slug: "custom",
		Permissions: []string{"content:read"}, Level: 50,
	}, 100)

	newName := "Custom Updated"
	newLevel := 55
	newPerms := []string{"content:read", "content:update"}
	updated, err := uc.Update(context.Background(), created.DocumentID, UpdateRoleInput{
		Name:        &newName,
		Level:       &newLevel,
		Permissions: &newPerms,
	}, 100)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "Custom Updated" {
		t.Errorf("Name = %q, want %q", updated.Name, "Custom Updated")
	}
	if updated.Level != 55 {
		t.Errorf("Level = %d, want 55", updated.Level)
	}
}

func TestDelete_DefaultRole(t *testing.T) {
	repo := seedRepo()
	userRepo := emptyUserRepo()
	uc := New(repo, userRepo)
	_ = uc.SeedDefaults(context.Background())

	editorRole, _ := repo.FindBySlug(context.Background(), "editor")
	err := uc.Delete(context.Background(), editorRole.DocumentID)
	if !pkgerrors.Is(err, pkgerrors.ErrValidation) {
		t.Errorf("err = %v, want ErrValidation for default role deletion", err)
	}
}

func TestDelete_CustomRole(t *testing.T) {
	repo := seedRepo()
	userRepo := emptyUserRepo()
	uc := New(repo, userRepo)

	created, _ := uc.Create(context.Background(), CreateRoleInput{
		Name: "Temp", Slug: "temp",
		Permissions: []string{"content:read"}, Level: 30,
	}, 100)

	if err := uc.Delete(context.Background(), created.DocumentID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.FindByID(context.Background(), created.DocumentID)
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestDelete_AssignedRole(t *testing.T) {
	repo := seedRepo()
	userRepo := &mock.UserRepository{
		FindAllFn: func(_ context.Context, _, _ int) ([]*entity.User, int64, error) {
			return []*entity.User{{ID: "u1", RoleID: "custom-doc-id"}}, 1, nil
		},
	}
	uc := New(repo, userRepo)

	created, _ := uc.Create(context.Background(), CreateRoleInput{
		Name: "InUse", Slug: "in-use",
		Permissions: []string{"content:read"}, Level: 30,
	}, 100)

	// Override the userRepo to simulate a user with this role
	userRepo.FindAllFn = func(_ context.Context, _, _ int) ([]*entity.User, int64, error) {
		return []*entity.User{{ID: "u1", RoleID: created.DocumentID}}, 1, nil
	}

	err := uc.Delete(context.Background(), created.DocumentID)
	if !pkgerrors.Is(err, pkgerrors.ErrConflict) {
		t.Errorf("err = %v, want ErrConflict for assigned role", err)
	}
}
