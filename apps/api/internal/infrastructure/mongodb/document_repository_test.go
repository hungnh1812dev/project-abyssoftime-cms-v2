//go:build !integration

package mongodb_test

import (
	"testing"

	"project-abyssoftime-cms-v2/api/internal/infrastructure/mongodb"
)

// compile-time: NewDocumentRepository must exist.
var _ = mongodb.NewDocumentRepository

func TestDocumentRepository_UpsertDraft_ThenFindDraftByEntryID(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestDocumentRepository_UpsertDraft_DifferentLocales_AreIsolated(t *testing.T) {
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

func TestDocumentRepository_DeletePublishedByEntryID_LeavesDraft(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestDocumentRepository_DeleteByContentType_RemovesAllEntries(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}
