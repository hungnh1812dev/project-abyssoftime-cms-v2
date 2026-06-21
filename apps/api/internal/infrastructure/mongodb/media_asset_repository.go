package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

var _ repository.MediaAssetRepository = (*mediaAssetRepository)(nil)

type mediaAssetRepository struct {
	col *mongo.Collection
}

func NewMediaAssetRepository(db *mongo.Database) repository.MediaAssetRepository {
	return &mediaAssetRepository{col: db.Collection("media_assets")}
}

func (r *mediaAssetRepository) Create(ctx context.Context, asset *entity.MediaAsset) error {
	if asset.ID == "" {
		asset.ID = primitive.NewObjectID().Hex()
	}
	if asset.CreatedAt.IsZero() {
		asset.CreatedAt = time.Now().UTC()
	}
	_, err := r.col.InsertOne(ctx, asset)
	return err
}

func (r *mediaAssetRepository) FindByID(ctx context.Context, id string) (*entity.MediaAsset, error) {
	var asset entity.MediaAsset
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&asset)
	if err == mongo.ErrNoDocuments {
		return nil, pkgerrors.ErrNotFound
	}
	return &asset, err
}

func (r *mediaAssetRepository) FindAll(ctx context.Context, page, limit int) ([]*entity.MediaAsset, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	skip := int64((page - 1) * limit)
	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(skip).
		SetLimit(int64(limit))

	total, err := r.col.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}
	cursor, err := r.col.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	var results []*entity.MediaAsset
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

func (r *mediaAssetRepository) FindByDocumentID(ctx context.Context, documentID string) (*entity.MediaAsset, error) {
	var asset entity.MediaAsset
	err := r.col.FindOne(ctx, bson.M{"documentId": documentID}).Decode(&asset)
	if err == mongo.ErrNoDocuments {
		return nil, pkgerrors.ErrNotFound
	}
	return &asset, err
}

func (r *mediaAssetRepository) Delete(ctx context.Context, id string) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return pkgerrors.ErrNotFound
	}
	return nil
}
