package entity

import "time"

type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleAdmin      Role = "admin"
	RoleEditor     Role = "editor"
	RoleGuest      Role = "guest"
)

func RoleLevel(role Role) int {
	switch role {
	case RoleSuperAdmin:
		return 4
	case RoleAdmin:
		return 3
	case RoleEditor:
		return 2
	case RoleGuest:
		return 1
	default:
		return 0
	}
}

type User struct {
	ID           uint      `bson:"_id,omitempty" gorm:"column:gorm_id;primaryKey;autoIncrement"`
	DocumentID   string    `bson:"documentId"    gorm:"column:document_id;uniqueIndex"`
	Email        string    `bson:"email"         gorm:"column:email;uniqueIndex"`
	DisplayName  string    `bson:"displayName"   gorm:"column:display_name"   json:"displayName"`
	PasswordHash string    `bson:"passwordHash"  gorm:"column:password_hash"`
	Role         Role      `bson:"role"          gorm:"column:role;type:varchar(20)"`
	RoleID       string    `bson:"roleId"        gorm:"column:role_id;index"`
	CreatedAt    time.Time `bson:"createdAt"     gorm:"column:created_at"`
}
