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
	ID            string          `bson:"_id,omitempty"`
	DocumentID    string          `bson:"documentId"`
	EntryID       string          `bson:"entryId"`
	Version       DocumentVersion `bson:"version"`
	ContentTypeID string          `bson:"contentTypeId"`
	Data          map[string]any  `bson:"data"`
	Locale        string          `bson:"locale"`
	CreatedAt     time.Time       `bson:"createdAt"`
	UpdatedAt     time.Time       `bson:"updatedAt"`
	PublishedAt   time.Time       `bson:"publishedAt,omitempty"`
	CreatedBy     string          `bson:"createdBy"`
	UpdatedBy     string          `bson:"updatedBy"`
	PublishedBy   string          `bson:"publishedBy,omitempty"`
}
