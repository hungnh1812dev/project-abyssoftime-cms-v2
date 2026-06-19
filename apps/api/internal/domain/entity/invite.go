package entity

import "time"

type Invite struct {
	ID        string    `bson:"_id,omitempty" gorm:"column:id;primaryKey"`
	Email     string    `bson:"email"         gorm:"column:email;uniqueIndex"`
	Role      Role      `bson:"role"          gorm:"column:role;type:varchar(20)"`
	TokenHash string    `bson:"tokenHash"     gorm:"column:token_hash;uniqueIndex"`
	ExpiresAt time.Time `bson:"expiresAt"     gorm:"column:expires_at"`
	CreatedBy string    `bson:"createdBy"     gorm:"column:created_by"`
	CreatedAt time.Time `bson:"createdAt"     gorm:"column:created_at"`
}
