package entity

import "time"

type Locale struct {
	ID        uint      `bson:"_id,omitempty"  gorm:"column:gorm_id;primaryKey;autoIncrement"  json:"-"`
	Code      string    `bson:"code"           gorm:"column:code;uniqueIndex"                  json:"code"`
	Name      string    `bson:"name"           gorm:"column:name"                              json:"name"`
	IsDefault bool      `bson:"isDefault"      gorm:"column:is_default"                        json:"isDefault"`
	CreatedAt time.Time `bson:"createdAt"      gorm:"column:created_at"                        json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt"      gorm:"column:updated_at"                        json:"updatedAt"`
}

func (Locale) TableName() string {
	return "locales"
}
