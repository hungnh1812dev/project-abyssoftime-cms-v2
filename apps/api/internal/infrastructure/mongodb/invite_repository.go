package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

var _ repository.InviteRepository = (*inviteRepository)(nil)

type inviteRepository struct {
	col *mongo.Collection
}

func NewInviteRepository(db *mongo.Database) repository.InviteRepository {
	return &inviteRepository{col: db.Collection("invites")}
}

func (r *inviteRepository) Create(ctx context.Context, invite *entity.Invite) error {

	if invite.CreatedAt.IsZero() {
		invite.CreatedAt = time.Now().UTC()
	}
	_, err := r.col.InsertOne(ctx, invite)
	if mongo.IsDuplicateKeyError(err) {
		return pkgerrors.ErrConflict
	}
	return err
}

func (r *inviteRepository) FindByHash(ctx context.Context, tokenHash string) (*entity.Invite, error) {
	var invite entity.Invite
	err := r.col.FindOne(ctx, bson.M{"tokenHash": tokenHash}).Decode(&invite)
	if err == mongo.ErrNoDocuments {
		return nil, pkgerrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (r *inviteRepository) FindByEmail(ctx context.Context, email string) (*entity.Invite, error) {
	var invite entity.Invite
	err := r.col.FindOne(ctx, bson.M{"email": email}).Decode(&invite)
	if err == mongo.ErrNoDocuments {
		return nil, pkgerrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (r *inviteRepository) Delete(ctx context.Context, id string) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return pkgerrors.ErrNotFound
	}
	return nil
}

func (r *inviteRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.col.DeleteMany(ctx, bson.M{"expiresAt": bson.M{"$lt": time.Now().UTC()}})
	return err
}

func (r *inviteRepository) FindAll(ctx context.Context) ([]*entity.Invite, error) {
	cursor, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var invites []*entity.Invite
	if err := cursor.All(ctx, &invites); err != nil {
		return nil, err
	}
	return invites, nil
}
