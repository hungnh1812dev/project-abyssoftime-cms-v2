package entity

import "testing"

func TestRoleLevel(t *testing.T) {
	tests := []struct {
		role Role
		want int
	}{
		{RoleSuperAdmin, 4},
		{RoleAdmin, 3},
		{RoleEditor, 2},
		{RoleGuest, 1},
		{Role("unknown"), 0},
		{Role(""), 0},
	}
	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			if got := RoleLevel(tt.role); got != tt.want {
				t.Errorf("RoleLevel(%q) = %d, want %d", tt.role, got, tt.want)
			}
		})
	}
}
