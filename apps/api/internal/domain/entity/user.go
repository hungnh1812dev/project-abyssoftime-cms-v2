package entity

import "time"

type Role string

const (
	RoleAdmin Role = "admin"
	RoleGuest Role = "guest"
)

type User struct {
	ID           string    `bson:"_id,omitempty" gorm:"column:id;primaryKey"`
	DocumentID   string    `bson:"documentId"    gorm:"column:document_id;uniqueIndex"`
	Email        string    `bson:"email"         gorm:"column:email;uniqueIndex"`
	PasswordHash string    `bson:"passwordHash"  gorm:"column:password_hash"`
	Role         Role      `bson:"role"          gorm:"column:role;type:varchar(20)"`
	CreatedAt    time.Time `bson:"createdAt"     gorm:"column:created_at"`
}
