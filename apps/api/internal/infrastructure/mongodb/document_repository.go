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
	if version == entity.VersionDraft && doc.GormID == 0 {
		col := r.collection(contentTypeSlug)
		var maxDoc struct {
			GormID uint `bson:"gormId"`
		}
		findOpts := options.FindOne().
			SetSort(bson.D{{Key: "gormId", Value: -1}}).
			SetProjection(bson.D{{Key: "gormId", Value: 1}})
		err := col.FindOne(ctx, bson.M{"version": entity.VersionDraft}, findOpts).Decode(&maxDoc)
		if err != nil && err != mongo.ErrNoDocuments {
			return err
		}
		doc.GormID = maxDoc.GormID + 1
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

var mongoBsonSortKey = map[string]string{
	"id":        "gormId",
	"createdAt": "createdAt",
	"updatedAt": "updatedAt",
}

func resolveBsonSortKey(orderBy string) string {
	if k, ok := mongoBsonSortKey[orderBy]; ok {
		return k
	}
	return "createdAt"
}

func (r *documentRepository) FindDraftsByContentTypePaginated(ctx context.Context, contentTypeSlug string, start, size int, locale, orderBy string, sortDir int) ([]*entity.Document, int64, error) {
	col := r.collection(contentTypeSlug)
	filter := bson.M{"version": entity.VersionDraft, "locale": locale}

	total, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: resolveBsonSortKey(orderBy), Value: sortDir}}).
		SetSkip(int64(start)).
		SetLimit(int64(size))
	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var results []*entity.Document
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

func (r *documentRepository) FindPublishedByContentTypePaginated(ctx context.Context, contentTypeSlug string, start, size int, locale, orderBy string, sortDir int) ([]*entity.Document, int64, error) {
	col := r.collection(contentTypeSlug)
	filter := bson.M{"version": entity.VersionPublished, "locale": locale}

	total, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: resolveBsonSortKey(orderBy), Value: sortDir}}).
		SetSkip(int64(start)).
		SetLimit(int64(size))
	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var results []*entity.Document
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

func (r *documentRepository) FindPublishedByDocumentIDs(ctx context.Context, contentTypeSlug string, documentIDs []string, locale string) ([]*entity.Document, error) {
	if len(documentIDs) == 0 {
		return nil, nil
	}
	filter := bson.M{
		"version":    entity.VersionPublished,
		"locale":     locale,
		"documentId": bson.M{"$in": documentIDs},
	}
	cursor, err := r.collection(contentTypeSlug).Find(ctx, filter)
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

func (r *documentRepository) EnsureCollection(ctx context.Context, contentTypeSlug string, _ []entity.FieldDefinition) error {
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

func (r *documentRepository) TableInfo(ctx context.Context, contentTypeSlug string) (bool, int64, error) {
	col := r.collection(contentTypeSlug)
	count, err := col.CountDocuments(ctx, bson.M{})
	if err != nil {
		return false, 0, err
	}
	return count > 0, count, nil
}
