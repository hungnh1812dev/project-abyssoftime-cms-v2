package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	indexes := []struct {
		collection string
		model      mongo.IndexModel
	}{
		{
			collection: "users",
			model: mongo.IndexModel{
				Keys:    bson.D{{Key: "email", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
		},
		{
			collection: "content_types",
			model: mongo.IndexModel{
				Keys:    bson.D{{Key: "slug", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
		},
		{
			collection: "documents",
			model: mongo.IndexModel{
				Keys: bson.D{{Key: "contentTypeId", Value: 1}},
			},
		},
		{
			collection: "media_assets",
			model: mongo.IndexModel{
				Keys: bson.D{{Key: "documentRef", Value: 1}},
			},
		},
	}

	for _, idx := range indexes {
		if _, err := db.Collection(idx.collection).Indexes().CreateOne(ctx, idx.model); err != nil {
			return err
		}
	}

	return nil
}
