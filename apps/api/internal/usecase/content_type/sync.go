package content_type

import (
	"context"
	"log"

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

	if exists, count, err := s.docRepo.TableInfo(ctx, def.Slug); err == nil {
		if exists {
			log.Printf("sync: %q — table exists, %d rows", def.Slug, count)
		} else {
			log.Printf("sync: %q — table not found, will create", def.Slug)
		}
	}

	current, err := s.FindBySlug(ctx, def.Slug)
	if err != nil {
		if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return err
		}
		ct := &entity.ContentType{Name: def.Name, Slug: def.Slug, Kind: kind, Fields: def.Fields}
		if err := s.Create(ctx, ct); err != nil {
			return err
		}
		if err := s.docRepo.EnsureCollection(ctx, def.Slug, def.Fields); err != nil {
			return err
		}
		return s.ensureComponentTables(ctx, def.Slug, def.Fields)
	}

	if err := s.docRepo.EnsureCollection(ctx, current.Slug, def.Fields); err != nil {
		return err
	}
	if err := s.ensureComponentTables(ctx, current.Slug, def.Fields); err != nil {
		return err
	}

	if current.Name == def.Name && current.Kind == kind && fieldsEqual(current.Fields, def.Fields) {
		return nil
	}
	current.Name = def.Name
	current.Kind = kind
	current.Fields = def.Fields
	return s.Update(ctx, current)
}

func (s *Syncer) ensureComponentTables(ctx context.Context, slug string, fields []entity.FieldDefinition) error {
	return s.ensureComponentTablesRecursive(ctx, slug, "", fields, 0)
}

func (s *Syncer) ensureComponentTablesRecursive(ctx context.Context, slug, prefix string, fields []entity.FieldDefinition, depth int) error {
	if s.compRepo == nil {
		return nil
	}
	for _, field := range fields {
		if field.Type == "component" {
			path := field.Name
			if prefix != "" {
				path = prefix + "_" + field.Name
			}
			isNested := depth > 0
			if err := s.compRepo.EnsureCollection(ctx, slug, path, field.Fields, isNested); err != nil {
				return err
			}
			if err := s.ensureComponentTablesRecursive(ctx, slug, path, field.Fields, depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Syncer) dropComponentTables(ctx context.Context, slug string, fields []entity.FieldDefinition) error {
	return s.dropComponentTablesRecursive(ctx, slug, "", fields)
}

func (s *Syncer) dropComponentTablesRecursive(ctx context.Context, slug, prefix string, fields []entity.FieldDefinition) error {
	for _, f := range fields {
		if f.Type == "component" {
			path := f.Name
			if prefix != "" {
				path = prefix + "_" + f.Name
			}
			if err := s.dropComponentTablesRecursive(ctx, slug, path, f.Fields); err != nil {
				return err
			}
			if err := s.compRepo.DropCollection(ctx, slug, path); err != nil {
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
		if a[i].Name != b[i].Name || a[i].Type != b[i].Type || a[i].Width != b[i].Width || a[i].Repeatable != b[i].Repeatable {
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
		if err := s.dropComponentTables(ctx, ct.Slug, ct.Fields); err != nil {
			return err
		}
	}
	if err := s.docRepo.DropCollection(ctx, ct.Slug); err != nil {
		return err
	}
	return s.UseCase.Delete(ctx, ct.DocumentID)
}
