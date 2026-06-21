package entity

import "time"

type ContentKind string

const (
	KindSingle     ContentKind = "single"
	KindCollection ContentKind = "collection"
)

type FieldDefinition struct {
	Name   string            `json:"name"             bson:"name"`
	Type   string            `json:"type"             bson:"type"`
	Ext    []string          `json:"ext,omitempty"    bson:"ext,omitempty"`
	Fields []FieldDefinition `json:"fields,omitempty" bson:"fields,omitempty"`
}

type ContentType struct {
	ID         uint              `bson:"_id,omitempty"            gorm:"column:gorm_id;primaryKey;autoIncrement"`
	DocumentID string            `bson:"documentId"               gorm:"column:document_id;uniqueIndex"`
	Name       string            `bson:"name"                     gorm:"column:name"`
	Slug       string            `bson:"slug"                     gorm:"column:slug;uniqueIndex"`
	Kind       ContentKind       `bson:"kind"                     gorm:"column:kind;type:varchar(20)"`
	Fields     []FieldDefinition `json:"Fields,omitempty"         bson:"fields,omitempty"     gorm:"column:fields;serializer:json"`
	ListFields []string          `json:"listFields,omitempty"     bson:"listFields,omitempty" gorm:"column:list_fields;serializer:json"`
	CreatedAt  time.Time         `bson:"createdAt"                gorm:"column:created_at"`
	UpdatedAt  time.Time         `bson:"updatedAt"                gorm:"column:updated_at"`
}
