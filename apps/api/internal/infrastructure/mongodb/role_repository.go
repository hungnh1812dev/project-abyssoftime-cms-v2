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

var _ repository.RoleRepository = (*roleRepository)(nil)

type roleRepository struct {
	col *mongo.Collection
}

func NewRoleRepository(db *mongo.Database) repository.RoleRepository {
	return &roleRepository{col: db.Collection("roles")}
}

func (r *roleRepository) Create(ctx context.Context, role *entity.RoleEntity) error {
	if role.ID == "" {
		role.ID = primitive.NewObjectID().Hex()
	}
	if role.CreatedAt.IsZero() {
		role.CreatedAt = time.Now().UTC()
	}
	if role.UpdatedAt.IsZero() {
		role.UpdatedAt = time.Now().UTC()
	}
	_, err := r.col.InsertOne(ctx, role)
	if mongo.IsDuplicateKeyError(err) {
		return pkgerrors.ErrConflict
	}
	return err
}

func (r *roleRepository) FindByID(ctx context.Context, documentID string) (*entity.RoleEntity, error) {
	var role entity.RoleEntity
	err := r.col.FindOne(ctx, bson.M{"documentId": documentID}).Decode(&role)
	if err == mongo.ErrNoDocuments {
		return nil, pkgerrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindBySlug(ctx context.Context, slug string) (*entity.RoleEntity, error) {
	var role entity.RoleEntity
	err := r.col.FindOne(ctx, bson.M{"slug": slug}).Decode(&role)
	if err == mongo.ErrNoDocuments {
		return nil, pkgerrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindAll(ctx context.Context) ([]*entity.RoleEntity, error) {
	opts := options.Find().SetSort(bson.D{{Key: "level", Value: -1}})
	cursor, err := r.col.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var roles []*entity.RoleEntity
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) Update(ctx context.Context, role *entity.RoleEntity) error {
	role.UpdatedAt = time.Now().UTC()
	res, err := r.col.ReplaceOne(ctx, bson.M{"documentId": role.DocumentID}, role)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return pkgerrors.ErrNotFound
	}
	return nil
}

func (r *roleRepository) Delete(ctx context.Context, documentID string) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"documentId": documentID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return pkgerrors.ErrNotFound
	}
	return nil
}

func (r *roleRepository) HasAny(ctx context.Context) (bool, error) {
	count, err := r.col.CountDocuments(ctx, bson.M{})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
