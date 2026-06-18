package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/handler"
	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

// ---- mock usecase ----------------------------------------------------------

type mockMediaUC struct {
	uploadFn   func(ctx context.Context, file io.Reader, filename, documentRef, contentTypeID string) (*entity.MediaAsset, error)
	listFn     func(ctx context.Context, page, limit int) ([]*entity.MediaAsset, int64, error)
	deleteFn   func(ctx context.Context, id string) error
}

func (m *mockMediaUC) Upload(ctx context.Context, file io.Reader, filename, documentRef, contentTypeID string) (*entity.MediaAsset, error) {
	return m.uploadFn(ctx, file, filename, documentRef, contentTypeID)
}

func (m *mockMediaUC) List(ctx context.Context, page, limit int) ([]*entity.MediaAsset, int64, error) {
	return m.listFn(ctx, page, limit)
}

func (m *mockMediaUC) Delete(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}

// ---- Upload ----------------------------------------------------------------

func TestMediaHandler_Upload_OK(t *testing.T) {
	uc := &mockMediaUC{}
	uc.uploadFn = func(_ context.Context, _ io.Reader, filename, _, _ string) (*entity.MediaAsset, error) {
		return &entity.MediaAsset{
			ID:  "asset-1",
			URL: "https://cdn.example.com/" + filename,
		}, nil
	}
	h := handler.NewMediaHandler(uc)

	body, contentType := buildMultipartForm(t, "photo.jpg", []byte("fake-image-data"), "doc-1", "ct-1")
	req := httptest.NewRequest(http.MethodPost, "/api/media/upload", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	h.Upload(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Upload() status = %d, want %d", w.Code, http.StatusCreated)
	}

	var respBody map[string]any
	if err := json.NewDecoder(w.Body).Decode(&respBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := respBody["url"]; !ok {
		t.Errorf("response JSON missing field %q; got keys: %v", "url", jsonKeys(respBody))
	}
	if _, ok := respBody["thumbnailUrl"]; !ok {
		t.Errorf("response JSON missing field %q; got keys: %v", "thumbnailUrl", jsonKeys(respBody))
	}
}

func jsonKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func TestMediaHandler_List_OK(t *testing.T) {
	uc := &mockMediaUC{}
	uc.listFn = func(_ context.Context, page, limit int) ([]*entity.MediaAsset, int64, error) {
		return []*entity.MediaAsset{
			{ID: "a1", URL: "https://cdn/a1.jpg"},
		}, 5, nil
	}
	h := handler.NewMediaHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/api/media?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("List() status = %d, want 200", w.Code)
	}
	var out map[string]any
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if _, ok := out["items"]; !ok {
		t.Error("List() response missing 'items' key")
	}
	if out["total"] != float64(5) {
		t.Errorf("List() total = %v, want 5", out["total"])
	}
	if out["page"] != float64(1) {
		t.Errorf("List() page = %v, want 1", out["page"])
	}
}

func TestMediaHandler_Upload_MissingFile_BadRequest(t *testing.T) {
	uc := &mockMediaUC{}
	h := handler.NewMediaHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/media/upload", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Upload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Upload() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// buildMultipartForm builds a multipart body with a file field and metadata fields.
func buildMultipartForm(t *testing.T, filename string, content []byte, documentRef, contentTypeID string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	fw.Write(content)

	w.WriteField("documentRef", documentRef)
	w.WriteField("contentTypeId", contentTypeID)
	w.Close()

	return &buf, w.FormDataContentType()
}

// ---- Delete ----------------------------------------------------------------

func serveDelete(t *testing.T, h *handler.MediaHandler, id string) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/media/{id}", h.Delete)
	req := httptest.NewRequest(http.MethodDelete, "/api/media/"+id, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}

func TestMediaHandler_Delete_Returns204(t *testing.T) {
	uc := &mockMediaUC{}
	uc.deleteFn = func(_ context.Context, _ string) error { return nil }
	h := handler.NewMediaHandler(uc)

	w := serveDelete(t, h, "asset-1")
	if w.Code != http.StatusNoContent {
		t.Errorf("Delete() status = %d, want 204", w.Code)
	}
}

func TestMediaHandler_Delete_NotFound_Returns404(t *testing.T) {
	uc := &mockMediaUC{}
	uc.deleteFn = func(_ context.Context, _ string) error {
		return pkgerrors.ErrNotFound
	}
	h := handler.NewMediaHandler(uc)

	w := serveDelete(t, h, "missing")
	if w.Code != http.StatusNotFound {
		t.Errorf("Delete() status = %d, want 404", w.Code)
	}
}

func TestMediaHandler_Delete_UseCaseError_Returns500(t *testing.T) {
	uc := &mockMediaUC{}
	uc.deleteFn = func(_ context.Context, _ string) error {
		return errors.New("storage error")
	}
	h := handler.NewMediaHandler(uc)

	w := serveDelete(t, h, "asset-1")
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Delete() status = %d, want 500", w.Code)
	}
}
