package user_test

import (
	"context"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	repomock "project-abyssoftime-cms-v2/api/internal/domain/repository/mock"
	"project-abyssoftime-cms-v2/api/internal/usecase/user"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

var testRoles = map[string]*entity.RoleEntity{
	"role-sa":     {DocumentID: "role-sa", Slug: "super_admin", Level: 100},
	"role-admin":  {DocumentID: "role-admin", Slug: "admin", Level: 80},
	"role-editor": {DocumentID: "role-editor", Slug: "editor", Level: 60},
	"role-guest":  {DocumentID: "role-guest", Slug: "guest", Level: 20},
}

func defaultRoleRepo() *repomock.RoleRepository {
	return &repomock.RoleRepository{
		FindByIDFn: func(_ context.Context, id string) (*entity.RoleEntity, error) {
			if r, ok := testRoles[id]; ok {
				return r, nil
			}
			return nil, pkgerrors.ErrNotFound
		},
		FindBySlugFn: func(_ context.Context, slug string) (*entity.RoleEntity, error) {
			for _, r := range testRoles {
				if r.Slug == slug {
					return r, nil
				}
			}
			return nil, pkgerrors.ErrNotFound
		},
		HasAnyFn: func(_ context.Context) (bool, error) { return true, nil },
	}
}

func makeUserRepo(users map[string]*entity.User) *repomock.UserRepository {
	return &repomock.UserRepository{
		FindByIDFn: func(_ context.Context, id string) (*entity.User, error) {
			u, ok := users[id]
			if !ok {
				return nil, pkgerrors.ErrNotFound
			}
			cp := *u
			return &cp, nil
		},
		UpdateFn: func(_ context.Context, u *entity.User) error {
			if _, ok := users[u.DocumentID]; !ok {
				return pkgerrors.ErrNotFound
			}
			users[u.DocumentID] = u
			return nil
		},
		DeleteFn: func(_ context.Context, id string) error {
			if _, ok := users[id]; !ok {
				return pkgerrors.ErrNotFound
			}
			delete(users, id)
			return nil
		},
		FindAllFn: func(_ context.Context, _, _ int) ([]*entity.User, int64, error) {
			var list []*entity.User
			for _, u := range users {
				list = append(list, u)
			}
			return list, int64(len(list)), nil
		},
	}
}

func TestUpdateRole(t *testing.T) {
	tests := []struct {
		name      string
		actor     *entity.User
		target    *entity.User
		newRoleID string
		wantErr   error
	}{
		{
			name:      "super_admin promotes guest to editor",
			actor:     &entity.User{DocumentID: "sa", RoleID: "role-sa"},
			target:    &entity.User{DocumentID: "g", RoleID: "role-guest"},
			newRoleID: "role-editor",
		},
		{
			name:      "super_admin promotes editor to admin",
			actor:     &entity.User{DocumentID: "sa", RoleID: "role-sa"},
			target:    &entity.User{DocumentID: "e", RoleID: "role-editor"},
			newRoleID: "role-admin",
		},
		{
			name:      "admin demotes editor to guest",
			actor:     &entity.User{DocumentID: "a", RoleID: "role-admin"},
			target:    &entity.User{DocumentID: "e", RoleID: "role-editor"},
			newRoleID: "role-guest",
		},
		{
			name:      "admin cannot change another admin",
			actor:     &entity.User{DocumentID: "a1", RoleID: "role-admin"},
			target:    &entity.User{DocumentID: "a2", RoleID: "role-admin"},
			newRoleID: "role-editor",
			wantErr:   pkgerrors.ErrForbidden,
		},
		{
			name:      "admin cannot promote to admin",
			actor:     &entity.User{DocumentID: "a", RoleID: "role-admin"},
			target:    &entity.User{DocumentID: "e", RoleID: "role-editor"},
			newRoleID: "role-admin",
			wantErr:   pkgerrors.ErrForbidden,
		},
		{
			name:      "editor cannot promote guest to editor (equal to own role)",
			actor:     &entity.User{DocumentID: "e", RoleID: "role-editor"},
			target:    &entity.User{DocumentID: "g", RoleID: "role-guest"},
			newRoleID: "role-editor",
			wantErr:   pkgerrors.ErrForbidden,
		},
		{
			name:      "cannot change own role",
			actor:     &entity.User{DocumentID: "sa", RoleID: "role-sa"},
			target:    &entity.User{DocumentID: "sa", RoleID: "role-sa"},
			newRoleID: "role-admin",
			wantErr:   pkgerrors.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := map[string]*entity.User{
				tt.actor.DocumentID: tt.actor,
			}
			if tt.actor.DocumentID != tt.target.DocumentID {
				users[tt.target.DocumentID] = tt.target
			}
			repo := makeUserRepo(users)
			uc := user.New(repo, defaultRoleRepo())

			err := uc.UpdateRole(context.Background(), tt.actor.DocumentID, tt.target.DocumentID, tt.newRoleID)
			if tt.wantErr != nil {
				if !pkgerrors.Is(err, tt.wantErr) {
					t.Errorf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if users[tt.target.DocumentID].RoleID != tt.newRoleID {
				t.Errorf("roleID = %q, want %q", users[tt.target.DocumentID].RoleID, tt.newRoleID)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name    string
		actor   *entity.User
		target  *entity.User
		wantErr error
	}{
		{
			name:   "super_admin deletes admin",
			actor:  &entity.User{DocumentID: "sa", RoleID: "role-sa"},
			target: &entity.User{DocumentID: "a", RoleID: "role-admin"},
		},
		{
			name:   "admin deletes editor",
			actor:  &entity.User{DocumentID: "a", RoleID: "role-admin"},
			target: &entity.User{DocumentID: "e", RoleID: "role-editor"},
		},
		{
			name:    "admin cannot delete another admin",
			actor:   &entity.User{DocumentID: "a1", RoleID: "role-admin"},
			target:  &entity.User{DocumentID: "a2", RoleID: "role-admin"},
			wantErr: pkgerrors.ErrForbidden,
		},
		{
			name:    "cannot delete self",
			actor:   &entity.User{DocumentID: "sa", RoleID: "role-sa"},
			target:  &entity.User{DocumentID: "sa", RoleID: "role-sa"},
			wantErr: pkgerrors.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := map[string]*entity.User{
				tt.actor.DocumentID: tt.actor,
			}
			if tt.actor.DocumentID != tt.target.DocumentID {
				users[tt.target.DocumentID] = tt.target
			}
			repo := makeUserRepo(users)
			uc := user.New(repo, defaultRoleRepo())

			err := uc.Delete(context.Background(), tt.actor.DocumentID, tt.target.DocumentID)
			if tt.wantErr != nil {
				if !pkgerrors.Is(err, tt.wantErr) {
					t.Errorf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if _, ok := users[tt.target.DocumentID]; ok {
				t.Error("target should have been deleted")
			}
		})
	}
}
