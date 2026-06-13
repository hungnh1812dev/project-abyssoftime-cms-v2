package entity

import "time"

type Role string

const (
	RoleAdmin Role = "admin"
	RoleGuest Role = "guest"
)

type User struct {
	ID           string    `bson:"_id,omitempty"`
	DocumentID   string    `bson:"documentId"`
	Email        string    `bson:"email"`
	PasswordHash string    `bson:"passwordHash"`
	Role         Role      `bson:"role"`
	CreatedAt    time.Time `bson:"createdAt"`
}
