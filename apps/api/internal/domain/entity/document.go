package entity

import "time"

type DocumentStatus string

const (
	StatusDraft     DocumentStatus = "draft"
	StatusPublished DocumentStatus = "published"
)

type Document struct {
	ID            string            `bson:"_id,omitempty"`
	DocumentID    string            `bson:"documentId"`
	ContentTypeID string            `bson:"contentTypeId"`
	Status        DocumentStatus    `bson:"status"`
	Data          map[string]any    `bson:"data"`
	CreatedAt     time.Time         `bson:"createdAt"`
	UpdatedAt     time.Time         `bson:"updatedAt"`
}
