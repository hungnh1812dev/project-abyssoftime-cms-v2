package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

var _ repository.LocaleRepository = (*localeRepository)(nil)

type localeRepository struct {
	col *mongo.Collection
}

func NewLocaleRepository(database *mongo.Database) repository.LocaleRepository {
	return &localeRepository{col: database.Collection("locales")}
}

func (repo *localeRepository) Create(ctx context.Context, locale *entity.Locale) error {
	_, err := repo.col.InsertOne(ctx, locale)
	if mongo.IsDuplicateKeyError(err) {
		return pkgerrors.ErrConflict
	}
	return err
}

func (repo *localeRepository) FindByCode(ctx context.Context, code string) (*entity.Locale, error) {
	var locale entity.Locale
	err := repo.col.FindOne(ctx, bson.M{"code": code}).Decode(&locale)
	if err == mongo.ErrNoDocuments {
		return nil, pkgerrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &locale, nil
}

func (repo *localeRepository) FindAll(ctx context.Context) ([]*entity.Locale, error) {
	opts := options.Find().SetSort(bson.D{{Key: "code", Value: 1}})
	cursor, err := repo.col.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var locales []*entity.Locale
	if err := cursor.All(ctx, &locales); err != nil {
		return nil, err
	}
	return locales, nil
}

func (repo *localeRepository) FindDefault(ctx context.Context) (*entity.Locale, error) {
	var locale entity.Locale
	err := repo.col.FindOne(ctx, bson.M{"isDefault": true}).Decode(&locale)
	if err == mongo.ErrNoDocuments {
		return nil, pkgerrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &locale, nil
}

func (repo *localeRepository) Update(ctx context.Context, locale *entity.Locale) error {
	res, err := repo.col.ReplaceOne(ctx, bson.M{"code": locale.Code}, locale)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return pkgerrors.ErrNotFound
	}
	return nil
}

func (repo *localeRepository) Delete(ctx context.Context, code string) error {
	res, err := repo.col.DeleteOne(ctx, bson.M{"code": code})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return pkgerrors.ErrNotFound
	}
	return nil
}

func (repo *localeRepository) ClearDefault(ctx context.Context) error {
	_, err := repo.col.UpdateMany(ctx, bson.M{"isDefault": true}, bson.M{"$set": bson.M{"isDefault": false}})
	return err
}
