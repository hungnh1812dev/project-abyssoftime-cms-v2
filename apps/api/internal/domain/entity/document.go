package entity

import "time"

type DocumentVersion string

const (
	VersionDraft     DocumentVersion = "draft"
	VersionPublished DocumentVersion = "published"
)

type Document struct {
	DocumentID    string          `bson:"documentId"                json:"documentId"              gorm:"column:document_id;index"`
	Version       DocumentVersion `bson:"version"                   json:"version"                 gorm:"column:version;type:varchar(20)"`
	ContentTypeID string          `bson:"contentTypeId"             json:"contentTypeId"           gorm:"column:content_type_id"`
	Data          map[string]any  `bson:"data"                      json:"data"                    gorm:"column:data;serializer:json"`
	Locale        string          `bson:"locale"                    json:"locale"                  gorm:"column:locale"`
	CreatedAt     time.Time       `bson:"createdAt"                 json:"createdAt"               gorm:"column:created_at"`
	UpdatedAt     time.Time       `bson:"updatedAt"                 json:"updatedAt"               gorm:"column:updated_at"`
	PublishedAt   time.Time       `bson:"publishedAt,omitempty"     json:"publishedAt,omitempty"   gorm:"column:published_at"`
	CreatedBy     string          `bson:"createdBy"                 json:"createdBy"               gorm:"column:created_by"`
	UpdatedBy     string          `bson:"updatedBy"                 json:"updatedBy"               gorm:"column:updated_by"`
	PublishedBy   string          `bson:"publishedBy,omitempty"     json:"publishedBy,omitempty"   gorm:"column:published_by"`
	Slug          string          `bson:"-"                         json:"-"                       gorm:"column:slug;index"`
}
