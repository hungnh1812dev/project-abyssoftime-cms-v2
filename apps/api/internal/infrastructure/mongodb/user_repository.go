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

var _ repository.UserRepository = (*userRepository)(nil)

type userRepository struct {
	col *mongo.Collection
}

func NewUserRepository(db *mongo.Database) repository.UserRepository {
	return &userRepository{col: db.Collection("users")}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}
	_, err := r.col.InsertOne(ctx, user)
	if mongo.IsDuplicateKeyError(err) {
		return pkgerrors.ErrConflict
	}
	return err
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.col.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, pkgerrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, pkgerrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByIDs(ctx context.Context, ids []string) ([]*entity.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	cursor, err := r.col.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*entity.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) HasSuperAdmin(ctx context.Context) (bool, error) {
	count, err := r.col.CountDocuments(ctx, bson.M{"role": string(entity.RoleSuperAdmin)})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userRepository) FindAll(ctx context.Context, page, limit int) ([]*entity.User, int64, error) {
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

	var users []*entity.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	res, err := r.col.ReplaceOne(ctx, bson.M{"_id": user.ID}, user)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return pkgerrors.ErrNotFound
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return pkgerrors.ErrNotFound
	}
	return nil
}
