package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/middleware"
	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type documentUseCase interface {
	Save(ctx context.Context, contentTypeSlug string, doc *entity.Document, userID string) (*entity.Document, error)
	GetForEdit(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, string, error)
	GetPublished(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error)
	GetAll(ctx context.Context, contentTypeSlug string) ([]*entity.Document, error)
	Publish(ctx context.Context, contentTypeSlug, documentID, locale, userID string) error
	Unpublish(ctx context.Context, contentTypeSlug, documentID, locale string) error
	Delete(ctx context.Context, contentTypeSlug, documentID string) error
}

type DocumentHandler struct {
	uc documentUseCase
}

func NewDocumentHandler(uc documentUseCase) *DocumentHandler {
	return &DocumentHandler{uc: uc}
}

type documentRequest struct {
	Data map[string]any `json:"data"`
}

type entrySummary struct {
	DocumentID    string
	ContentTypeID string
	Data          map[string]any
	Status        string
	Locale        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CreatedBy     string
	UpdatedBy     string
}

func localeParam(r *http.Request) string {
	return r.URL.Query().Get("locale")
}

func toSummary(doc *entity.Document, status string) entrySummary {
	return entrySummary{
		DocumentID:    doc.DocumentID,
		ContentTypeID: doc.ContentTypeID,
		Data:          doc.Data,
		Status:        status,
		Locale:        doc.Locale,
		CreatedAt:     doc.CreatedAt,
		UpdatedAt:     doc.UpdatedAt,
		CreatedBy:     doc.CreatedBy,
		UpdatedBy:     doc.UpdatedBy,
	}
}

func (h *DocumentHandler) List(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	drafts, err := h.uc.GetAll(r.Context(), slug)
	if err != nil {
		writeErr(w, err)
		return
	}
	locale := localeParam(r)
	summaries := make([]entrySummary, 0, len(drafts))
	for _, draft := range drafts {
		_, status, err := h.uc.GetForEdit(r.Context(), slug, draft.DocumentID, locale)
		if err != nil {
			writeErr(w, err)
			return
		}
		summaries = append(summaries, toSummary(draft, status))
	}
	writeJSON(w, http.StatusOK, summaries)
}

func (h *DocumentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req documentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	slug := r.PathValue("slug")
	doc := &entity.Document{Data: req.Data, Locale: localeParam(r)}
	saved, err := h.uc.Save(r.Context(), slug, doc, middleware.UserID(r.Context()))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toSummary(saved, "draft"))
}

func (h *DocumentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	documentID := r.PathValue("documentId")
	draft, status, err := h.uc.GetForEdit(r.Context(), slug, documentID, localeParam(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toSummary(draft, status))
}

func (h *DocumentHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	documentID := r.PathValue("documentId")
	doc, err := h.uc.GetPublished(r.Context(), slug, documentID, localeParam(r))
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
	slug := r.PathValue("slug")
	documentID := r.PathValue("documentId")
	doc := &entity.Document{
		DocumentID: documentID,
		Data:       req.Data,
		Locale:     localeParam(r),
	}
	saved, err := h.uc.Save(r.Context(), slug, doc, middleware.UserID(r.Context()))
	if err != nil {
		writeErr(w, err)
		return
	}
	_, status, err := h.uc.GetForEdit(r.Context(), slug, saved.DocumentID, saved.Locale)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toSummary(saved, status))
}

func (h *DocumentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	documentID := r.PathValue("documentId")
	if err := h.uc.Delete(r.Context(), slug, documentID); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *DocumentHandler) Publish(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	documentID := r.PathValue("documentId")
	if err := h.uc.Publish(r.Context(), slug, documentID, localeParam(r), middleware.UserID(r.Context())); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "published"})
}

func (h *DocumentHandler) Unpublish(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	documentID := r.PathValue("documentId")
	if err := h.uc.Unpublish(r.Context(), slug, documentID, localeParam(r)); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "draft"})
}
