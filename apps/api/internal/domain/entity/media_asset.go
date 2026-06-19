package entity

import "time"

type MediaAsset struct {
	ID            string    `bson:"_id,omitempty"  json:"ID"            gorm:"column:id;primaryKey"`
	DocumentID    string    `bson:"documentId"     json:"documentId"    gorm:"column:document_id"`
	URL           string    `bson:"url"            json:"url"           gorm:"column:url"`
	ThumbnailURL  string    `bson:"thumbnailUrl"   json:"thumbnailUrl"  gorm:"column:thumbnail_url"`
	PublicID      string    `bson:"publicId"       json:"publicId"      gorm:"column:public_id"`
	FileName      string    `bson:"fileName"       json:"fileName"      gorm:"column:file_name"`
	FileExt       string    `bson:"fileExt"        json:"fileExt"       gorm:"column:file_ext"`
	Hash          string    `bson:"hash"           json:"hash"          gorm:"column:hash"`
	ContentTypeID string    `bson:"contentTypeId"  json:"contentTypeId" gorm:"column:content_type_id"`
	DocumentRef   string    `bson:"documentRef"    json:"documentRef"   gorm:"column:document_ref;index"`
	CreatedAt     time.Time `bson:"createdAt"      json:"createdAt"     gorm:"column:created_at"`
}
