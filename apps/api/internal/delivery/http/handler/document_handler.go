package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type documentUseCase interface {
	Create(ctx context.Context, doc *entity.Document) error
	GetOne(ctx context.Context, id string) (*entity.Document, error)
	GetAll(ctx context.Context, contentTypeID string) ([]*entity.Document, error)
	Update(ctx context.Context, doc *entity.Document) error
	Delete(ctx context.Context, id string) error
	Publish(ctx context.Context, id string) error
	Unpublish(ctx context.Context, id string) error
}

type DocumentHandler struct {
	uc documentUseCase
}

func NewDocumentHandler(uc documentUseCase) *DocumentHandler {
	return &DocumentHandler{uc: uc}
}

type documentRequest struct {
	ContentTypeID string         `json:"contentTypeId"`
	Data          map[string]any `json:"data"`
}

func (h *DocumentHandler) List(w http.ResponseWriter, r *http.Request) {
	contentTypeID := r.URL.Query().Get("contentType")
	docs, err := h.uc.GetAll(r.Context(), contentTypeID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, docs)
}

func (h *DocumentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req documentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	doc := &entity.Document{
		ContentTypeID: req.ContentTypeID,
		Data:          req.Data,
	}
	if err := h.uc.Create(r.Context(), doc); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, doc)
}

func (h *DocumentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	doc, err := h.uc.GetOne(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

func (h *DocumentHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req documentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	doc := &entity.Document{
		ID:            r.PathValue("id"),
		ContentTypeID: req.ContentTypeID,
		Data:          req.Data,
	}
	if err := h.uc.Update(r.Context(), doc); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

func (h *DocumentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *DocumentHandler) Publish(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Publish(r.Context(), r.PathValue("id")); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": string(entity.StatusPublished)})
}

func (h *DocumentHandler) Unpublish(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Unpublish(r.Context(), r.PathValue("id")); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": string(entity.StatusDraft)})
}
