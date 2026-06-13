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

var _ repository.DocumentRepository = (*documentRepository)(nil)

type documentRepository struct {
	col *mongo.Collection
}

func NewDocumentRepository(db *mongo.Database) repository.DocumentRepository {
	return &documentRepository{col: db.Collection("documents")}
}

func (r *documentRepository) Create(ctx context.Context, doc *entity.Document) error {
	if doc.ID == "" {
		doc.ID = primitive.NewObjectID().Hex()
	}
	now := time.Now().UTC()
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = now
	}
	doc.UpdatedAt = now
	if doc.Status == "" {
		doc.Status = entity.StatusDraft
	}
	_, err := r.col.InsertOne(ctx, doc)
	return err
}

func (r *documentRepository) FindByID(ctx context.Context, id string) (*entity.Document, error) {
	var doc entity.Document
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, pkgerrors.ErrNotFound
	}
	return &doc, err
}

func (r *documentRepository) FindByContentType(ctx context.Context, contentTypeID string) ([]*entity.Document, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := r.col.Find(ctx, bson.M{"contentTypeId": contentTypeID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var results []*entity.Document
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *documentRepository) Update(ctx context.Context, doc *entity.Document) error {
	doc.UpdatedAt = time.Now().UTC()
	res, err := r.col.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return pkgerrors.ErrNotFound
	}
	return nil
}

func (r *documentRepository) UpdateStatus(ctx context.Context, id string, status entity.DocumentStatus) error {
	res, err := r.col.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"status": status, "updatedAt": time.Now().UTC()}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return pkgerrors.ErrNotFound
	}
	return nil
}

func (r *documentRepository) Delete(ctx context.Context, id string) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return pkgerrors.ErrNotFound
	}
	return nil
}
