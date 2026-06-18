package entity

import "time"

type MediaAsset struct {
	ID            string    `bson:"_id,omitempty"  json:"ID"`
	DocumentID    string    `bson:"documentId"     json:"documentId"`
	URL           string    `bson:"url"            json:"url"`
	ThumbnailURL  string    `bson:"thumbnailUrl"   json:"thumbnailUrl"`
	PublicID      string    `bson:"publicId"       json:"publicId"`
	FileName      string    `bson:"fileName"       json:"fileName"`
	FileExt       string    `bson:"fileExt"        json:"fileExt"`
	Hash          string    `bson:"hash"           json:"hash"`
	ContentTypeID string    `bson:"contentTypeId"  json:"contentTypeId"`
	DocumentRef   string    `bson:"documentRef"    json:"documentRef"`
	CreatedAt     time.Time `bson:"createdAt"      json:"createdAt"`
}
