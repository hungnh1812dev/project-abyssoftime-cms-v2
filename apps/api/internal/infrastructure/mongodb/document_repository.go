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
	db *mongo.Database
}

func NewDocumentRepository(db *mongo.Database) repository.DocumentRepository {
	return &documentRepository{db: db}
}

func (r *documentRepository) collection(contentTypeSlug string) *mongo.Collection {
	return r.db.Collection("documents_" + contentTypeSlug)
}

func versionFilter(documentID string, version entity.DocumentVersion, locale string) bson.M {
	return bson.M{"documentId": documentID, "version": version, "locale": locale}
}

func documentLocaleFilter(documentID, locale string) bson.M {
	return bson.M{"documentId": documentID, "locale": locale}
}

func (r *documentRepository) findByDocumentAndVersion(ctx context.Context, contentTypeSlug, documentID, locale string, version entity.DocumentVersion) (*entity.Document, error) {
	var doc entity.Document
	err := r.collection(contentTypeSlug).FindOne(ctx, versionFilter(documentID, version, locale)).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, pkgerrors.ErrNotFound
	}
	return &doc, err
}

func (r *documentRepository) FindDraftByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error) {
	return r.findByDocumentAndVersion(ctx, contentTypeSlug, documentID, locale, entity.VersionDraft)
}

func (r *documentRepository) FindPublishedByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error) {
	return r.findByDocumentAndVersion(ctx, contentTypeSlug, documentID, locale, entity.VersionPublished)
}

func (r *documentRepository) upsertVersion(ctx context.Context, contentTypeSlug string, doc *entity.Document, version entity.DocumentVersion) error {
	doc.Version = version
	if doc.DocumentID == "" {
		doc.DocumentID = primitive.NewObjectID().Hex()
	}
	_, err := r.collection(contentTypeSlug).ReplaceOne(
		ctx,
		versionFilter(doc.DocumentID, version, doc.Locale),
		doc,
		options.Replace().SetUpsert(true),
	)
	return err
}

func (r *documentRepository) UpsertDraft(ctx context.Context, contentTypeSlug string, doc *entity.Document) error {
	return r.upsertVersion(ctx, contentTypeSlug, doc, entity.VersionDraft)
}

func (r *documentRepository) UpsertPublished(ctx context.Context, contentTypeSlug string, doc *entity.Document) error {
	return r.upsertVersion(ctx, contentTypeSlug, doc, entity.VersionPublished)
}

func (r *documentRepository) FindDraftsByContentType(ctx context.Context, contentTypeSlug string) ([]*entity.Document, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := r.collection(contentTypeSlug).Find(ctx, bson.M{"version": entity.VersionDraft}, opts)
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

func (r *documentRepository) DeleteByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) error {
	_, err := r.collection(contentTypeSlug).DeleteMany(ctx, documentLocaleFilter(documentID, locale))
	return err
}

func (r *documentRepository) DeletePublishedByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) error {
	_, err := r.collection(contentTypeSlug).DeleteOne(ctx, versionFilter(documentID, entity.VersionPublished, locale))
	return err
}

func (r *documentRepository) DeleteAllByContentType(ctx context.Context, contentTypeSlug string) error {
	_, err := r.collection(contentTypeSlug).DeleteMany(ctx, bson.M{})
	return err
}

func (r *documentRepository) EnsureCollection(ctx context.Context, contentTypeSlug string) error {
	col := r.collection(contentTypeSlug)
	_, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "documentId", Value: 1}, {Key: "version", Value: 1}, {Key: "locale", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (r *documentRepository) DropCollection(ctx context.Context, contentTypeSlug string) error {
	return r.collection(contentTypeSlug).Drop(ctx)
}
