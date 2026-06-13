package entity

import "time"

type MediaAsset struct {
	ID            string    `bson:"_id,omitempty"`
	DocumentID    string    `bson:"documentId"`
	URL           string    `bson:"url"`
	PublicID      string    `bson:"publicId"`
	ContentTypeID string    `bson:"contentTypeId"`
	DocumentRef   string    `bson:"documentRef"`
	CreatedAt     time.Time `bson:"createdAt"`
}
