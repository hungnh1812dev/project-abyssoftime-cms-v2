package handler

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type mediaUseCase interface {
	Upload(ctx context.Context, file io.Reader, filename, documentRef, contentTypeID string) (*entity.MediaAsset, error)
	List(ctx context.Context, page, limit int) ([]*entity.MediaAsset, int64, error)
	Delete(ctx context.Context, id string) error
}

type MediaHandler struct {
	uc mediaUseCase
}

func NewMediaHandler(uc mediaUseCase) *MediaHandler {
	return &MediaHandler{uc: uc}
}

func (h *MediaHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	items, total, err := h.uc.List(r.Context(), page, limit)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": items,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *MediaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.uc.Delete(r.Context(), id); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file field is required")
		return
	}
	defer file.Close()

	documentRef := r.FormValue("documentRef")
	contentTypeID := r.FormValue("contentTypeId")

	asset, err := h.uc.Upload(r.Context(), file, header.Filename, documentRef, contentTypeID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, asset)
}
