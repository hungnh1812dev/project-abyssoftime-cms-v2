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

var _ repository.AccessTokenRepository = (*accessTokenRepository)(nil)

type accessTokenRepository struct {
	col *mongo.Collection
}

func NewAccessTokenRepository(db *mongo.Database) repository.AccessTokenRepository {
	return &accessTokenRepository{col: db.Collection("access_tokens")}
}

func (r *accessTokenRepository) Create(ctx context.Context, token *entity.AccessToken) error {
	if token.ID == "" {
		token.ID = primitive.NewObjectID().Hex()
	}
	if token.CreatedAt.IsZero() {
		token.CreatedAt = time.Now().UTC()
	}
	_, err := r.col.InsertOne(ctx, token)
	if mongo.IsDuplicateKeyError(err) {
		return pkgerrors.ErrConflict
	}
	return err
}

func (r *accessTokenRepository) FindByHash(ctx context.Context, tokenHash string) (*entity.AccessToken, error) {
	var token entity.AccessToken
	err := r.col.FindOne(ctx, bson.M{"tokenHash": tokenHash}).Decode(&token)
	if err == mongo.ErrNoDocuments {
		return nil, pkgerrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *accessTokenRepository) FindAll(ctx context.Context, page, limit int) ([]*entity.AccessToken, int64, error) {
	total, err := r.col.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, err := r.col.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var tokens []*entity.AccessToken
	if err := cursor.All(ctx, &tokens); err != nil {
		return nil, 0, err
	}
	return tokens, total, nil
}

func (r *accessTokenRepository) Delete(ctx context.Context, id string) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return pkgerrors.ErrNotFound
	}
	return nil
}

func (r *accessTokenRepository) UpdateLastUsed(ctx context.Context, id string, at time.Time) error {
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"lastUsedAt": at}})
	return err
}
