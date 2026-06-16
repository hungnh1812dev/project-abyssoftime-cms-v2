package handler

import (
	"context"
	"net/http"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type contentTypeUseCase interface {
	FindByID(ctx context.Context, id string) (*entity.ContentType, error)
	FindAll(ctx context.Context) ([]*entity.ContentType, error)
}

type ContentTypeHandler struct {
	uc contentTypeUseCase
}

func NewContentTypeHandler(uc contentTypeUseCase) *ContentTypeHandler {
	return &ContentTypeHandler{uc: uc}
}

func (h *ContentTypeHandler) List(w http.ResponseWriter, r *http.Request) {
	cts, err := h.uc.FindAll(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cts)
}

func (h *ContentTypeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ct, err := h.uc.FindByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ct)
}
