package entity

import "time"

type Component struct {
	GormID      uint            `gorm:"column:gorm_id;primaryKey;autoIncrement"`
	ComponentID string          `gorm:"column:component_id"`
	DocumentID  string          `gorm:"column:document_id"`
	Version     DocumentVersion `gorm:"column:version;type:varchar(20)"`
	Locale      string          `gorm:"column:locale"`
	SortOrder   int             `gorm:"column:sort_order"`
	Fields      map[string]any  `gorm:"-"`
	CreatedAt   time.Time       `gorm:"column:created_at"`
	UpdatedAt   time.Time       `gorm:"column:updated_at"`
}
