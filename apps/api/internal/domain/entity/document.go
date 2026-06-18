package entity

import "time"

// DocumentVersion distinguishes the two physical records (in the same
// collection) that make up one logical entry: its draft and its published
// snapshot.
type DocumentVersion string

const (
	VersionDraft     DocumentVersion = "draft"
	VersionPublished DocumentVersion = "published"
)

type Document struct {
	DocumentID    string          `bson:"documentId" json:"documentId"`
	Version       DocumentVersion `bson:"version" json:"version"`
	ContentTypeID string          `bson:"contentTypeId" json:"contentTypeId"`
	Data          map[string]any  `bson:"data" json:"data"`
	Locale        string          `bson:"locale" json:"locale"`
	CreatedAt     time.Time       `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time       `bson:"updatedAt" json:"updatedAt"`
	PublishedAt   time.Time       `bson:"publishedAt,omitempty" json:"publishedAt,omitempty"`
	CreatedBy     string          `bson:"createdBy" json:"createdBy"`
	UpdatedBy     string          `bson:"updatedBy" json:"updatedBy"`
	PublishedBy   string          `bson:"publishedBy,omitempty" json:"publishedBy,omitempty"`
}
