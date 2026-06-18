package entity

import "time"

type ContentKind string

const (
	KindSingle     ContentKind = "single"
	KindCollection ContentKind = "collection"
)

type FieldDefinition struct {
	Name   string            `json:"name"            bson:"name"`
	Type   string            `json:"type"            bson:"type"`
	Ext    []string          `json:"ext,omitempty"   bson:"ext,omitempty"`
	Fields []FieldDefinition `json:"fields,omitempty" bson:"fields,omitempty"`
}

type ContentType struct {
	ID         string            `bson:"_id,omitempty"`
	DocumentID string            `bson:"documentId"`
	Name       string            `bson:"name"`
	Slug       string            `bson:"slug"`
	Kind       ContentKind       `bson:"kind"`
	Fields     []FieldDefinition `json:"Fields,omitempty" bson:"fields,omitempty"`
	CreatedAt  time.Time         `bson:"createdAt"`
	UpdatedAt  time.Time         `bson:"updatedAt"`
}
