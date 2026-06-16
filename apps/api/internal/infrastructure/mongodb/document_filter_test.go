package mongodb

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

func TestVersionFilter_IncludesLocale(t *testing.T) {
	got := versionFilter("entry-1", entity.VersionDraft, "vi")
	want := bson.M{"entryId": "entry-1", "version": entity.VersionDraft, "locale": "vi"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("versionFilter() = %v, want %v", got, want)
	}
}

func TestEntryLocaleFilter_ScopesToLocale(t *testing.T) {
	got := entryLocaleFilter("entry-1", "en")
	want := bson.M{"entryId": "entry-1", "locale": "en"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("entryLocaleFilter() = %v, want %v", got, want)
	}
}
