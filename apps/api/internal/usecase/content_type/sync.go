package content_type

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

// EntryDeleter is the subset of the document usecase that Sync needs to
// cascade-delete entries belonging to a content type removed from the
// JSON definitions.
type EntryDeleter interface {
	GetAll(ctx context.Context, contentTypeID string) ([]*entity.Document, error)
	Delete(ctx context.Context, id string) error
}

// Syncer reconciles content-type JSON definitions (the source of truth)
// against the ContentType records in MongoDB.
type Syncer struct {
	*UseCase
	entries EntryDeleter
}

func NewSyncer(uc *UseCase, entries EntryDeleter) *Syncer {
	return &Syncer{UseCase: uc, entries: entries}
}

// Sync creates ContentTypes for new definitions, updates changed ones in
// place, and deletes (cascading to their entries) any ContentType whose
// definition file no longer exists.
func (s *Syncer) Sync(ctx context.Context, defs []ContentTypeDefinition) error {
	existing, err := s.FindAll(ctx)
	if err != nil {
		return err
	}

	defSlugs := make(map[string]bool, len(defs))
	for _, def := range defs {
		defSlugs[def.Slug] = true
		if err := s.syncOne(ctx, def); err != nil {
			return err
		}
	}

	for _, ct := range existing {
		if defSlugs[ct.Slug] {
			continue
		}
		if err := s.removeContentType(ctx, ct); err != nil {
			return err
		}
	}

	return nil
}

func (s *Syncer) syncOne(ctx context.Context, def ContentTypeDefinition) error {
	kind := entity.ContentKind(def.Kind)

	current, err := s.FindBySlug(ctx, def.Slug)
	if err != nil {
		if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return err
		}
		return s.Create(ctx, &entity.ContentType{Name: def.Name, Slug: def.Slug, Kind: kind})
	}

	if current.Name == def.Name && current.Kind == kind {
		return nil
	}
	current.Name = def.Name
	current.Kind = kind
	return s.Update(ctx, current)
}

func (s *Syncer) removeContentType(ctx context.Context, ct *entity.ContentType) error {
	docs, err := s.entries.GetAll(ctx, ct.ID)
	if err != nil {
		return err
	}
	for _, doc := range docs {
		if err := s.entries.Delete(ctx, doc.ID); err != nil {
			return err
		}
	}
	return s.Delete(ctx, ct.ID)
}
