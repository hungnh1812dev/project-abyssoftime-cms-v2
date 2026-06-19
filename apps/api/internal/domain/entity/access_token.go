package entity

import "time"

type AccessToken struct {
	ID         string     `bson:"_id,omitempty" gorm:"column:id;primaryKey"`
	Name       string     `bson:"name"          gorm:"column:name"`
	TokenHash  string     `bson:"tokenHash"     gorm:"column:token_hash;uniqueIndex"`
	Prefix     string     `bson:"prefix"        gorm:"column:prefix"`
	Scopes     []string   `bson:"scopes"        gorm:"column:scopes;serializer:json"`
	ExpiresAt  *time.Time `bson:"expiresAt"     gorm:"column:expires_at"`
	LastUsedAt *time.Time `bson:"lastUsedAt"    gorm:"column:last_used_at"`
	CreatedBy  string     `bson:"createdBy"     gorm:"column:created_by"`
	CreatedAt  time.Time  `bson:"createdAt"     gorm:"column:created_at"`
}
