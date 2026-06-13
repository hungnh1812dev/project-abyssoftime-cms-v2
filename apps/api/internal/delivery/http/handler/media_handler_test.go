package handler_test

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/handler"
	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

// ---- mock usecase ----------------------------------------------------------

type mockMediaUC struct {
	uploadFn func(ctx context.Context, file io.Reader, filename, documentRef, contentTypeID string) (*entity.MediaAsset, error)
}

func (m *mockMediaUC) Upload(ctx context.Context, file io.Reader, filename, documentRef, contentTypeID string) (*entity.MediaAsset, error) {
	return m.uploadFn(ctx, file, filename, documentRef, contentTypeID)
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
