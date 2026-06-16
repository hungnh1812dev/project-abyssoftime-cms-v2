package mongodb

import (
	"context"

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

func (r *documentRepository) findByEntryAndVersion(ctx context.Context, entryID string, version entity.DocumentVersion) (*entity.Document, error) {
	var doc entity.Document
	err := r.col.FindOne(ctx, bson.M{"entryId": entryID, "version": version}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, pkgerrors.ErrNotFound
	}
	return &doc, err
}

func (r *documentRepository) FindDraftByEntryID(ctx context.Context, entryID string) (*entity.Document, error) {
	return r.findByEntryAndVersion(ctx, entryID, entity.VersionDraft)
}

func (r *documentRepository) FindPublishedByEntryID(ctx context.Context, entryID string) (*entity.Document, error) {
	return r.findByEntryAndVersion(ctx, entryID, entity.VersionPublished)
}

func (r *documentRepository) upsertVersion(ctx context.Context, doc *entity.Document, version entity.DocumentVersion) error {
	doc.Version = version
	if doc.ID == "" {
		doc.ID = primitive.NewObjectID().Hex()
	}
	_, err := r.col.ReplaceOne(
		ctx,
		bson.M{"entryId": doc.EntryID, "version": version},
		doc,
		options.Replace().SetUpsert(true),
	)
	return err
}

func (r *documentRepository) UpsertDraft(ctx context.Context, doc *entity.Document) error {
	return r.upsertVersion(ctx, doc, entity.VersionDraft)
}

func (r *documentRepository) UpsertPublished(ctx context.Context, doc *entity.Document) error {
	return r.upsertVersion(ctx, doc, entity.VersionPublished)
}

func (r *documentRepository) FindEntryDraftsByContentType(ctx context.Context, contentTypeID string) ([]*entity.Document, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := r.col.Find(ctx, bson.M{"contentTypeId": contentTypeID, "version": entity.VersionDraft}, opts)
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

func (r *documentRepository) DeleteByEntryID(ctx context.Context, entryID string) error {
	_, err := r.col.DeleteMany(ctx, bson.M{"entryId": entryID})
	return err
}

func (r *documentRepository) DeletePublishedByEntryID(ctx context.Context, entryID string) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"entryId": entryID, "version": entity.VersionPublished})
	return err
}

func (r *documentRepository) DeleteByContentType(ctx context.Context, contentTypeID string) error {
	_, err := r.col.DeleteMany(ctx, bson.M{"contentTypeId": contentTypeID})
	return err
}
