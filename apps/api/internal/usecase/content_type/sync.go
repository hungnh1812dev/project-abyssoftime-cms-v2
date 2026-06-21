package content_type

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

// EntryManager is the subset of the document usecase that Sync needs:
// cascade-deleting entries belonging to a content type removed from the
// JSON definitions. GetAll returns each entry's draft record; Delete
// takes a documentID (not a single record's own Mongo _id).
type EntryManager interface {
	GetAll(ctx context.Context, contentTypeSlug string) ([]*entity.Document, error)
	Delete(ctx context.Context, contentTypeSlug, documentID string, fields []entity.FieldDefinition) error
}

type Syncer struct {
	*UseCase
	entries  EntryManager
	docRepo  repository.DocumentRepository
	compRepo repository.ComponentRepository
}

func NewSyncer(uc *UseCase, entries EntryManager, docRepo repository.DocumentRepository, compRepo repository.ComponentRepository) *Syncer {
	return &Syncer{UseCase: uc, entries: entries, docRepo: docRepo, compRepo: compRepo}
}

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
		ct := &entity.ContentType{Name: def.Name, Slug: def.Slug, Kind: kind, Fields: def.Fields, ListFields: def.ListFields}
		if err := s.Create(ctx, ct); err != nil {
			return err
		}
		if err := s.docRepo.EnsureCollection(ctx, def.Slug); err != nil {
			return err
		}
		return s.ensureComponentTables(ctx, def.Slug, def.Fields)
	}

	if err := s.docRepo.EnsureCollection(ctx, current.Slug); err != nil {
		return err
	}
	if err := s.ensureComponentTables(ctx, current.Slug, def.Fields); err != nil {
		return err
	}

	if current.Name == def.Name && current.Kind == kind && fieldsEqual(current.Fields, def.Fields) && stringSliceEqual(current.ListFields, def.ListFields) {
		return nil
	}
	current.Name = def.Name
	current.Kind = kind
	current.Fields = def.Fields
	current.ListFields = def.ListFields
	return s.Update(ctx, current)
}

func (s *Syncer) ensureComponentTables(ctx context.Context, slug string, fields []entity.FieldDefinition) error {
	if s.compRepo == nil {
		return nil
	}
	for _, f := range fields {
		if f.Type == "component" {
			if err := s.compRepo.EnsureCollection(ctx, slug, f.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func fieldsEqual(a, b []entity.FieldDefinition) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name != b[i].Name || a[i].Type != b[i].Type {
			return false
		}
		if !fieldsEqual(a[i].Fields, b[i].Fields) {
			return false
		}
	}
	return true
}

func (s *Syncer) removeContentType(ctx context.Context, ct *entity.ContentType) error {
	docs, err := s.entries.GetAll(ctx, ct.Slug)
	if err != nil {
		return err
	}
	for _, doc := range docs {
		if err := s.entries.Delete(ctx, ct.Slug, doc.DocumentID, ct.Fields); err != nil {
			return err
		}
	}
	if s.compRepo != nil {
		for _, f := range ct.Fields {
			if f.Type == "component" {
				if err := s.compRepo.DropCollection(ctx, ct.Slug, f.Name); err != nil {
					return err
				}
			}
		}
	}
	if err := s.docRepo.DropCollection(ctx, ct.Slug); err != nil {
		return err
	}
	return s.UseCase.Delete(ctx, ct.ID)
}
