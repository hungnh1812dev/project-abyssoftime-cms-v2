package entity

import "time"

type MediaAsset struct {
	ID           uint      `bson:"_id,omitempty"  json:"ID"            gorm:"column:gorm_id;primaryKey;autoIncrement"`
	DocumentID   string    `bson:"documentId"     json:"documentId"    gorm:"column:document_id;uniqueIndex"`
	URL          string    `bson:"url"            json:"url"           gorm:"column:url"`
	ThumbnailURL string    `bson:"thumbnailUrl"   json:"thumbnailUrl"  gorm:"column:thumbnail_url"`
	PublicID     string    `bson:"publicId"       json:"publicId"      gorm:"column:public_id"`
	FileName     string    `bson:"fileName"       json:"fileName"      gorm:"column:file_name"`
	FileExt      string    `bson:"fileExt"        json:"fileExt"       gorm:"column:file_ext"`
	Hash         string    `bson:"hash"           json:"hash"          gorm:"column:hash"`
	Width        int       `bson:"width"          json:"width"         gorm:"column:width"`
	Height       int       `bson:"height"         json:"height"        gorm:"column:height"`
	CreatedAt    time.Time `bson:"createdAt"      json:"createdAt"     gorm:"column:created_at"`
}
