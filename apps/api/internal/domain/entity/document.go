package entity

import "time"

// Deprecated: superseded by computing status from a DocumentVersion pair
// (see Domain Rules: Draft & Publish in SPEC.md). Removed once the document
// usecase is migrated to the entry-aware repository methods.
type DocumentStatus string

const (
	StatusDraft     DocumentStatus = "draft"
	StatusPublished DocumentStatus = "published"
)

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
	Status        DocumentStatus  `bson:"status"`
	Data          map[string]any  `bson:"data"`
	Locale        string          `bson:"locale"`
	CreatedAt     time.Time       `bson:"createdAt"`
	UpdatedAt     time.Time       `bson:"updatedAt"`
	PublishedAt   time.Time       `bson:"publishedAt,omitempty"`
	CreatedBy     string          `bson:"createdBy"`
	UpdatedBy     string          `bson:"updatedBy"`
	PublishedBy   string          `bson:"publishedBy,omitempty"`
}
