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

	"github.com/gin-gonic/gin"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/handler"
	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

type mockMediaUC struct {
	uploadFn func(ctx context.Context, file io.Reader, filename string) (*entity.MediaAsset, error)
	listFn   func(ctx context.Context, page, limit int) ([]*entity.MediaAsset, int64, error)
	deleteFn func(ctx context.Context, id string) error
}

func (m *mockMediaUC) Upload(ctx context.Context, file io.Reader, filename string) (*entity.MediaAsset, error) {
	return m.uploadFn(ctx, file, filename)
}

func (m *mockMediaUC) List(ctx context.Context, page, limit int) ([]*entity.MediaAsset, int64, error) {
	return m.listFn(ctx, page, limit)
}

func (m *mockMediaUC) Delete(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}

func buildMultipartForm(t *testing.T, filename string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	fw.Write(content)
	w.Close()

	return &buf, w.FormDataContentType()
}

func jsonKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func TestMediaHandler_Upload_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockMediaUC{}
	uc.uploadFn = func(_ context.Context, _ io.Reader, filename string) (*entity.MediaAsset, error) {
		return &entity.MediaAsset{
			DocumentID: "asset-1",
			URL: "https://cdn.example.com/" + filename,
		}, nil
	}
	h := handler.NewMediaHandler(uc)

	body, contentType := buildMultipartForm(t, "photo.jpg", []byte("fake-image-data"))
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/api/media/upload", h.Upload)

	req := httptest.NewRequest(http.MethodPost, "/api/media/upload", body)
	req.Header.Set("Content-Type", contentType)
	r.ServeHTTP(w, req)

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

func TestMediaHandler_List_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockMediaUC{}
	uc.listFn = func(_ context.Context, page, limit int) ([]*entity.MediaAsset, int64, error) {
		return []*entity.MediaAsset{
			{DocumentID: "a1", URL: "https://cdn/a1.jpg"},
		}, 5, nil
	}
	h := handler.NewMediaHandler(uc)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/api/media", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/media?page=1&limit=10", nil)
	r.ServeHTTP(w, req)

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
	gin.SetMode(gin.TestMode)

	uc := &mockMediaUC{}
	h := handler.NewMediaHandler(uc)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/api/media/upload", h.Upload)

	req := httptest.NewRequest(http.MethodPost, "/api/media/upload", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Upload() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func ginServeDelete(t *testing.T, h *handler.MediaHandler, id string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.DELETE("/api/media/:id", h.Delete)
	req := httptest.NewRequest(http.MethodDelete, "/api/media/"+id, nil)
	r.ServeHTTP(w, req)
	return w
}

func TestMediaHandler_Delete_Returns204(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockMediaUC{}
	uc.deleteFn = func(_ context.Context, _ string) error { return nil }
	h := handler.NewMediaHandler(uc)

	w := ginServeDelete(t, h, "asset-1")
	if w.Code != http.StatusNoContent {
		t.Errorf("Delete() status = %d, want 204", w.Code)
	}
}

func TestMediaHandler_Delete_NotFound_Returns404(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockMediaUC{}
	uc.deleteFn = func(_ context.Context, _ string) error {
		return pkgerrors.ErrNotFound
	}
	h := handler.NewMediaHandler(uc)

	w := ginServeDelete(t, h, "missing")
	if w.Code != http.StatusNotFound {
		t.Errorf("Delete() status = %d, want 404", w.Code)
	}
}

func TestMediaHandler_Delete_UseCaseError_Returns500(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockMediaUC{}
	uc.deleteFn = func(_ context.Context, _ string) error {
		return errors.New("storage error")
	}
	h := handler.NewMediaHandler(uc)

	w := ginServeDelete(t, h, "asset-1")
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Delete() status = %d, want 500", w.Code)
	}
}
