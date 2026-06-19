package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGinWriteErr(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
	}{
		{
			name:       "ErrConflict maps to 409",
			err:        fmt.Errorf("duplicate: %w", pkgerrors.ErrConflict),
			wantStatus: http.StatusConflict,
			wantMsg:    "duplicate: conflict",
		},
		{
			name:       "ErrUnauthorized maps to 401",
			err:        fmt.Errorf("bad token: %w", pkgerrors.ErrUnauthorized),
			wantStatus: http.StatusUnauthorized,
			wantMsg:    "bad token: unauthorized",
		},
		{
			name:       "ErrForbidden maps to 403",
			err:        fmt.Errorf("no access: %w", pkgerrors.ErrForbidden),
			wantStatus: http.StatusForbidden,
			wantMsg:    "no access: forbidden",
		},
		{
			name:       "ErrNotFound maps to 404",
			err:        fmt.Errorf("missing: %w", pkgerrors.ErrNotFound),
			wantStatus: http.StatusNotFound,
			wantMsg:    "missing: not found",
		},
		{
			name:       "ErrBadRequest maps to 400",
			err:        fmt.Errorf("invalid: %w", pkgerrors.ErrBadRequest),
			wantStatus: http.StatusBadRequest,
			wantMsg:    "invalid: bad request",
		},
		{
			name:       "ErrValidation maps to 422",
			err:        fmt.Errorf("failed: %w", pkgerrors.ErrValidation),
			wantStatus: http.StatusUnprocessableEntity,
			wantMsg:    "failed: validation error",
		},
		{
			name:       "unknown error maps to 500 with generic message",
			err:        fmt.Errorf("something broke"),
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			ginWriteErr(c, tt.err)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			var body map[string]string
			if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["error"] != tt.wantMsg {
				t.Errorf("error = %q, want %q", body["error"], tt.wantMsg)
			}
		})
	}
}

func TestGinWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ginWriteError(c, http.StatusBadRequest, "invalid input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["error"] != "invalid input" {
		t.Errorf("error = %q, want %q", body["error"], "invalid input")
	}
}
