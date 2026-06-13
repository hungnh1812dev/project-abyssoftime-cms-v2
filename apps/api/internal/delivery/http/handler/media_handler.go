package handler

import (
	"context"
	"io"
	"net/http"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type mediaUseCase interface {
	Upload(ctx context.Context, file io.Reader, filename, documentRef, contentTypeID string) (*entity.MediaAsset, error)
}

type MediaHandler struct {
	uc mediaUseCase
}

func NewMediaHandler(uc mediaUseCase) *MediaHandler {
	return &MediaHandler{uc: uc}
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
