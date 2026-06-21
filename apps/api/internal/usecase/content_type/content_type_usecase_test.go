package content_type_test

import (
	"context"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	repomock "project-abyssoftime-cms-v2/api/internal/domain/repository/mock"
	contenttype "project-abyssoftime-cms-v2/api/internal/usecase/content_type"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

var ctx = context.Background()

// ---- Create ----------------------------------------------------------------

func TestCreate(t *testing.T) {
	tests := []struct {
		name      string
		input     *entity.ContentType
		setupRepo func(*repomock.ContentTypeRepository)
		wantErr   error
	}{
		{
			name:  "valid single-type create",
			input: &entity.ContentType{Name: "Homepage", Slug: "homepage", Kind: entity.KindSingle},
			setupRepo: func(r *repomock.ContentTypeRepository) {
				r.FindBySlugFn = func(_ context.Context, _ string) (*entity.ContentType, error) {
					return nil, pkgerrors.ErrNotFound
				}
				r.CreateFn = func(_ context.Context, ct *entity.ContentType) error {
					return nil
				}
			},
			wantErr: nil,
		},
		{
			name:  "valid collection-type create",
			input: &entity.ContentType{Name: "Blog", Slug: "blog", Kind: entity.KindCollection},
			setupRepo: func(r *repomock.ContentTypeRepository) {
				r.FindBySlugFn = func(_ context.Context, _ string) (*entity.ContentType, error) {
					return nil, pkgerrors.ErrNotFound
				}
				r.CreateFn = func(_ context.Context, ct *entity.ContentType) error {
					return nil
				}
			},
			wantErr: nil,
		},
		{
			name:  "duplicate slug → conflict",
			input: &entity.ContentType{Name: "Homepage", Slug: "homepage", Kind: entity.KindSingle},
			setupRepo: func(r *repomock.ContentTypeRepository) {
				r.FindBySlugFn = func(_ context.Context, _ string) (*entity.ContentType, error) {
					return &entity.ContentType{DocumentID: "existing-id"}, nil
				}
			},
			wantErr: pkgerrors.ErrConflict,
		},
		{
			name:  "invalid kind → bad request",
			input: &entity.ContentType{Name: "Bad", Slug: "bad", Kind: "unknown"},
			setupRepo: func(r *repomock.ContentTypeRepository) {
				// FindBySlug not called when kind is invalid
			},
			wantErr: pkgerrors.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repomock.ContentTypeRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}
			uc := contenttype.New(repo)
			err := uc.Create(ctx, tt.input)
			if !pkgerrors.Is(err, tt.wantErr) {
				t.Errorf("Create() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && tt.input.DocumentID == "" {
				t.Error("Create() did not set ID on ContentType")
			}
		})
	}
}

// ---- FindByID --------------------------------------------------------------

func TestFindByID(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		setupRepo func(*repomock.ContentTypeRepository)
		wantErr   error
		wantID    string
	}{
		{
			name: "found",
			id:   "abc",
			setupRepo: func(r *repomock.ContentTypeRepository) {
				r.FindByIDFn = func(_ context.Context, id string) (*entity.ContentType, error) {
					return &entity.ContentType{DocumentID: id}, nil
				}
			},
			wantID: "abc",
		},
		{
			name: "not found",
			id:   "missing",
			setupRepo: func(r *repomock.ContentTypeRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.ContentType, error) {
					return nil, pkgerrors.ErrNotFound
				}
			},
			wantErr: pkgerrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repomock.ContentTypeRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}
			uc := contenttype.New(repo)
			ct, err := uc.FindByID(ctx, tt.id)
			if !pkgerrors.Is(err, tt.wantErr) {
				t.Errorf("FindByID() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantID != "" && (ct == nil || ct.DocumentID != tt.wantID) {
				t.Errorf("FindByID() ID = %v, want %v", ct, tt.wantID)
			}
		})
	}
}

// ---- FindAll ---------------------------------------------------------------

func TestFindAll(t *testing.T) {
	repo := &repomock.ContentTypeRepository{}
	repo.FindAllFn = func(_ context.Context) ([]*entity.ContentType, error) {
		return []*entity.ContentType{
			{DocumentID: "1", Slug: "blog"},
			{DocumentID: "2", Slug: "homepage"},
		}, nil
	}
	uc := contenttype.New(repo)
	list, err := uc.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}
	if len(list) != 2 {
		t.Errorf("FindAll() count = %d, want 2", len(list))
	}
}

// ---- Update ----------------------------------------------------------------

func TestUpdate(t *testing.T) {
	tests := []struct {
		name      string
		input     *entity.ContentType
		setupRepo func(*repomock.ContentTypeRepository)
		wantErr   error
	}{
		{
			name:  "valid update",
			input: &entity.ContentType{DocumentID: "abc", Name: "Updated", Slug: "homepage", Kind: entity.KindSingle},
			setupRepo: func(r *repomock.ContentTypeRepository) {
				r.UpdateFn = func(_ context.Context, ct *entity.ContentType) error { return nil }
			},
			wantErr: nil,
		},
		{
			name:  "invalid kind → bad request",
			input: &entity.ContentType{DocumentID: "abc", Kind: "bad"},
			setupRepo: func(r *repomock.ContentTypeRepository) {},
			wantErr:   pkgerrors.ErrBadRequest,
		},
		{
			name:  "not found",
			input: &entity.ContentType{DocumentID: "missing", Kind: entity.KindCollection},
			setupRepo: func(r *repomock.ContentTypeRepository) {
				r.UpdateFn = func(_ context.Context, _ *entity.ContentType) error {
					return pkgerrors.ErrNotFound
				}
			},
			wantErr: pkgerrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repomock.ContentTypeRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}
			uc := contenttype.New(repo)
			err := uc.Update(ctx, tt.input)
			if !pkgerrors.Is(err, tt.wantErr) {
				t.Errorf("Update() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// ---- Delete ----------------------------------------------------------------

func TestDelete(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		setupRepo func(*repomock.ContentTypeRepository)
		wantErr   error
	}{
		{
			name: "ok",
			id:   "abc",
			setupRepo: func(r *repomock.ContentTypeRepository) {
				r.DeleteFn = func(_ context.Context, _ string) error { return nil }
			},
			wantErr: nil,
		},
		{
			name: "not found",
			id:   "missing",
			setupRepo: func(r *repomock.ContentTypeRepository) {
				r.DeleteFn = func(_ context.Context, _ string) error { return pkgerrors.ErrNotFound }
			},
			wantErr: pkgerrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repomock.ContentTypeRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}
			uc := contenttype.New(repo)
			err := uc.Delete(ctx, tt.id)
			if !pkgerrors.Is(err, tt.wantErr) {
				t.Errorf("Delete() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
