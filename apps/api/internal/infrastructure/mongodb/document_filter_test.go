package mongodb

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

func TestVersionFilter_IncludesLocale(t *testing.T) {
	got := versionFilter("doc-1", entity.VersionDraft, "vi")
	want := bson.M{"documentId": "doc-1", "version": entity.VersionDraft, "locale": "vi"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("versionFilter() = %v, want %v", got, want)
	}
}

func TestDocumentLocaleFilter_ScopesToLocale(t *testing.T) {
	got := documentLocaleFilter("doc-1", "en")
	want := bson.M{"documentId": "doc-1", "locale": "en"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("documentLocaleFilter() = %v, want %v", got, want)
	}
}
