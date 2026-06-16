package content_type

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

// EntryManager is the subset of the document usecase that Sync needs:
// auto-creating a new single-type's singleton entry, and cascade-deleting
// entries belonging to a content type removed from the JSON definitions.
// GetAll returns each entry's draft record; Delete takes an EntryID (not a
// single record's own Mongo _id).
type EntryManager interface {
	Save(ctx context.Context, doc *entity.Document, userID string) (*entity.Document, error)
	GetAll(ctx context.Context, contentTypeID string) ([]*entity.Document, error)
	Delete(ctx context.Context, entryID string) error
}

// syncUser is the audit identity recorded on entries the sync step creates
// itself (e.g. a single-type's singleton), rather than a real editor.
const syncUser = "system"

// Syncer reconciles content-type JSON definitions (the source of truth)
// against the ContentType records in MongoDB.
type Syncer struct {
	*UseCase
	entries EntryManager
}

func NewSyncer(uc *UseCase, entries EntryManager) *Syncer {
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
		ct := &entity.ContentType{Name: def.Name, Slug: def.Slug, Kind: kind}
		if err := s.Create(ctx, ct); err != nil {
			return err
		}
		if kind == entity.KindSingle {
			return s.createSingleton(ctx, ct)
		}
		return nil
	}

	if current.Name == def.Name && current.Kind == kind {
		return nil
	}
	current.Name = def.Name
	current.Kind = kind
	return s.Update(ctx, current)
}

// createSingleton creates the one-and-only entry for a newly-created
// single-type content type, with entryID = contentTypeID per the
// Single-Type domain rule.
func (s *Syncer) createSingleton(ctx context.Context, ct *entity.ContentType) error {
	doc := &entity.Document{EntryID: ct.ID, ContentTypeID: ct.ID, Data: map[string]any{}}
	_, err := s.entries.Save(ctx, doc, syncUser)
	return err
}

func (s *Syncer) removeContentType(ctx context.Context, ct *entity.ContentType) error {
	docs, err := s.entries.GetAll(ctx, ct.ID)
	if err != nil {
		return err
	}
	for _, doc := range docs {
		if err := s.entries.Delete(ctx, doc.EntryID); err != nil {
			return err
		}
	}
	return s.Delete(ctx, ct.ID)
}
