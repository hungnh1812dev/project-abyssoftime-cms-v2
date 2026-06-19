package errors_test

import (
	"fmt"
	"testing"

	pkgerr "project-abyssoftime-cms-v2/api/pkg/errors"
)

func TestSentinelErrors(t *testing.T) {
	errs := []error{
		pkgerr.ErrNotFound,
		pkgerr.ErrUnauthorized,
		pkgerr.ErrForbidden,
		pkgerr.ErrConflict,
		pkgerr.ErrBadRequest,
		pkgerr.ErrValidation,
	}
	for _, e := range errs {
		if e == nil {
			t.Fatalf("sentinel error must not be nil")
		}
	}
}

func TestIs(t *testing.T) {
	wrapped := fmt.Errorf("repo: %w", pkgerr.ErrNotFound)
	if !pkgerr.Is(wrapped, pkgerr.ErrNotFound) {
		t.Fatalf("Is should unwrap wrapped errors")
	}
}
