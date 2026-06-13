package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type contentTypeUseCase interface {
	Create(ctx context.Context, ct *entity.ContentType) error
	FindByID(ctx context.Context, id string) (*entity.ContentType, error)
	FindAll(ctx context.Context) ([]*entity.ContentType, error)
	Update(ctx context.Context, ct *entity.ContentType) error
	Delete(ctx context.Context, id string) error
}

type ContentTypeHandler struct {
	uc contentTypeUseCase
}

func NewContentTypeHandler(uc contentTypeUseCase) *ContentTypeHandler {
	return &ContentTypeHandler{uc: uc}
}

type contentTypeRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	Kind string `json:"kind"`
}

func (h *ContentTypeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req contentTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	ct := &entity.ContentType{
		Name: req.Name,
		Slug: req.Slug,
		Kind: entity.ContentKind(req.Kind),
	}
	if err := h.uc.Create(r.Context(), ct); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, ct)
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

func (h *ContentTypeHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req contentTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	ct := &entity.ContentType{
		ID:   r.PathValue("id"),
		Name: req.Name,
		Slug: req.Slug,
		Kind: entity.ContentKind(req.Kind),
	}
	if err := h.uc.Update(r.Context(), ct); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ct)
}

func (h *ContentTypeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
