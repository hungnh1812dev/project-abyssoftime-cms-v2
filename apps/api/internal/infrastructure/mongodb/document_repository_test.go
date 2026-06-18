//go:build !integration

package mongodb_test

import (
	"testing"

	"project-abyssoftime-cms-v2/api/internal/infrastructure/mongodb"
)

// compile-time: NewDocumentRepository must exist.
var _ = mongodb.NewDocumentRepository

func TestDocumentRepository_UpsertDraft_ThenFindDraftByDocumentID(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestDocumentRepository_UpsertDraft_DifferentLocales_AreIsolated(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestDocumentRepository_FindPublishedByDocumentID_NotFoundUntilPublished(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestDocumentRepository_FindDraftsByContentType_OneRowPerEntry(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestDocumentRepository_DeleteByDocumentID_RemovesDraftAndPublished(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestDocumentRepository_DeletePublishedByDocumentID_LeavesDraft(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestDocumentRepository_DeleteAllByContentType_RemovesAllEntries(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}
