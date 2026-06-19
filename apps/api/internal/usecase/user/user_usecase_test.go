package user_test

import (
	"context"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	repomock "project-abyssoftime-cms-v2/api/internal/domain/repository/mock"
	"project-abyssoftime-cms-v2/api/internal/usecase/user"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

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
			if _, ok := users[u.ID]; !ok {
				return pkgerrors.ErrNotFound
			}
			users[u.ID] = u
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
		name    string
		actor   *entity.User
		target  *entity.User
		newRole entity.Role
		wantErr error
	}{
		{
			name:    "super_admin promotes guest to editor",
			actor:   &entity.User{ID: "sa", Role: entity.RoleSuperAdmin},
			target:  &entity.User{ID: "g", Role: entity.RoleGuest},
			newRole: entity.RoleEditor,
		},
		{
			name:    "super_admin promotes editor to admin",
			actor:   &entity.User{ID: "sa", Role: entity.RoleSuperAdmin},
			target:  &entity.User{ID: "e", Role: entity.RoleEditor},
			newRole: entity.RoleAdmin,
		},
		{
			name:    "admin demotes editor to guest",
			actor:   &entity.User{ID: "a", Role: entity.RoleAdmin},
			target:  &entity.User{ID: "e", Role: entity.RoleEditor},
			newRole: entity.RoleGuest,
		},
		{
			name:    "admin cannot change another admin",
			actor:   &entity.User{ID: "a1", Role: entity.RoleAdmin},
			target:  &entity.User{ID: "a2", Role: entity.RoleAdmin},
			newRole: entity.RoleEditor,
			wantErr: pkgerrors.ErrForbidden,
		},
		{
			name:    "admin cannot promote to admin",
			actor:   &entity.User{ID: "a", Role: entity.RoleAdmin},
			target:  &entity.User{ID: "e", Role: entity.RoleEditor},
			newRole: entity.RoleAdmin,
			wantErr: pkgerrors.ErrForbidden,
		},
		{
			name:    "editor cannot promote guest to editor (equal to own role)",
			actor:   &entity.User{ID: "e", Role: entity.RoleEditor},
			target:  &entity.User{ID: "g", Role: entity.RoleGuest},
			newRole: entity.RoleEditor,
			wantErr: pkgerrors.ErrForbidden,
		},
		{
			name:    "cannot change own role",
			actor:   &entity.User{ID: "sa", Role: entity.RoleSuperAdmin},
			target:  &entity.User{ID: "sa", Role: entity.RoleSuperAdmin},
			newRole: entity.RoleAdmin,
			wantErr: pkgerrors.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := map[string]*entity.User{
				tt.actor.ID: tt.actor,
			}
			if tt.actor.ID != tt.target.ID {
				users[tt.target.ID] = tt.target
			}
			repo := makeUserRepo(users)
			uc := user.New(repo)

			err := uc.UpdateRole(context.Background(), tt.actor.ID, tt.target.ID, tt.newRole)
			if tt.wantErr != nil {
				if !pkgerrors.Is(err, tt.wantErr) {
					t.Errorf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if users[tt.target.ID].Role != tt.newRole {
				t.Errorf("role = %q, want %q", users[tt.target.ID].Role, tt.newRole)
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
			actor:  &entity.User{ID: "sa", Role: entity.RoleSuperAdmin},
			target: &entity.User{ID: "a", Role: entity.RoleAdmin},
		},
		{
			name:   "admin deletes editor",
			actor:  &entity.User{ID: "a", Role: entity.RoleAdmin},
			target: &entity.User{ID: "e", Role: entity.RoleEditor},
		},
		{
			name:    "admin cannot delete another admin",
			actor:   &entity.User{ID: "a1", Role: entity.RoleAdmin},
			target:  &entity.User{ID: "a2", Role: entity.RoleAdmin},
			wantErr: pkgerrors.ErrForbidden,
		},
		{
			name:    "cannot delete self",
			actor:   &entity.User{ID: "sa", Role: entity.RoleSuperAdmin},
			target:  &entity.User{ID: "sa", Role: entity.RoleSuperAdmin},
			wantErr: pkgerrors.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := map[string]*entity.User{
				tt.actor.ID: tt.actor,
			}
			if tt.actor.ID != tt.target.ID {
				users[tt.target.ID] = tt.target
			}
			repo := makeUserRepo(users)
			uc := user.New(repo)

			err := uc.Delete(context.Background(), tt.actor.ID, tt.target.ID)
			if tt.wantErr != nil {
				if !pkgerrors.Is(err, tt.wantErr) {
					t.Errorf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if _, ok := users[tt.target.ID]; ok {
				t.Error("target should have been deleted")
			}
		})
	}
}
