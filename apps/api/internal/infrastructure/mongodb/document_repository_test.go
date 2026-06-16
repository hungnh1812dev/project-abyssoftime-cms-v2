//go:build !integration

package mongodb_test

import (
	"context"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/infrastructure/mongodb"
)

// compile-time: NewDocumentRepository must exist.
var _ = mongodb.NewDocumentRepository

func TestDocumentRepository_Create_SetsID(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestDocumentRepository_FindByID_NotFound(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestDocumentRepository_FindByContentType_Empty(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestDocumentRepository_UpdateStatus(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestDocumentRepository_Delete_NotFound(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestDocumentRepository_UpsertDraft_ThenFindDraftByEntryID(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestDocumentRepository_FindPublishedByEntryID_NotFoundUntilPublished(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestDocumentRepository_FindEntryDraftsByContentType_OneRowPerEntry(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestDocumentRepository_DeleteByEntryID_RemovesDraftAndPublished(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestDocumentRepository_DeleteByContentType_RemovesAllEntries(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

var documentRepoTests = []struct {
	name          string
	contentTypeID string
	status        entity.DocumentStatus
}{
	{"draft doc", "ct-1", entity.StatusDraft},
	{"published doc", "ct-1", entity.StatusPublished},
}

func TestDocumentRepository_TableDriven(t *testing.T) {
	_ = context.Background
	_ = documentRepoTests
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}
