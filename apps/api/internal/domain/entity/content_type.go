package entity

import "time"

type ContentKind string

const (
	KindSingle     ContentKind = "single"
	KindCollection ContentKind = "collection"
)

type ContentType struct {
	ID         string      `bson:"_id,omitempty"`
	DocumentID string      `bson:"documentId"`
	Name       string      `bson:"name"`
	Slug       string      `bson:"slug"`
	Kind       ContentKind `bson:"kind"`
	CreatedAt  time.Time   `bson:"createdAt"`
	UpdatedAt  time.Time   `bson:"updatedAt"`
}
