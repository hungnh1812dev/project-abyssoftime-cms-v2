package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

var _ repository.ContentTypeRepository = (*contentTypeRepository)(nil)

type contentTypeRepository struct {
	col *mongo.Collection
}

func NewContentTypeRepository(db *mongo.Database) repository.ContentTypeRepository {
	return &contentTypeRepository{col: db.Collection("content_types")}
}

func (r *contentTypeRepository) Create(ctx context.Context, ct *entity.ContentType) error {

	now := time.Now().UTC()
	if ct.CreatedAt.IsZero() {
		ct.CreatedAt = now
	}
	ct.UpdatedAt = now
	_, err := r.col.InsertOne(ctx, ct)
	if mongo.IsDuplicateKeyError(err) {
		return pkgerrors.ErrConflict
	}
	return err
}

func (r *contentTypeRepository) FindByID(ctx context.Context, id string) (*entity.ContentType, error) {
	var ct entity.ContentType
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&ct)
	if err == mongo.ErrNoDocuments {
		return nil, pkgerrors.ErrNotFound
	}
	return &ct, err
}

func (r *contentTypeRepository) FindBySlug(ctx context.Context, slug string) (*entity.ContentType, error) {
	var ct entity.ContentType
	err := r.col.FindOne(ctx, bson.M{"slug": slug}).Decode(&ct)
	if err == mongo.ErrNoDocuments {
		return nil, pkgerrors.ErrNotFound
	}
	return &ct, err
}

func (r *contentTypeRepository) FindAll(ctx context.Context) ([]*entity.ContentType, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}})
	cursor, err := r.col.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var results []*entity.ContentType
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *contentTypeRepository) Update(ctx context.Context, ct *entity.ContentType) error {
	ct.UpdatedAt = time.Now().UTC()
	res, err := r.col.ReplaceOne(ctx, bson.M{"_id": ct.DocumentID}, ct)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return pkgerrors.ErrNotFound
	}
	return nil
}

func (r *contentTypeRepository) Delete(ctx context.Context, id string) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return pkgerrors.ErrNotFound
	}
	return nil
}
