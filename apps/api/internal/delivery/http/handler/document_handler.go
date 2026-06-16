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
	Save(ctx context.Context, doc *entity.Document, userID string) (*entity.Document, error)
	GetForEdit(ctx context.Context, entryID string) (*entity.Document, string, error)
	GetPublished(ctx context.Context, entryID string) (*entity.Document, error)
	GetAll(ctx context.Context, contentTypeID string) ([]*entity.Document, error)
	Publish(ctx context.Context, entryID, userID string) error
	Unpublish(ctx context.Context, entryID string) error
	Delete(ctx context.Context, entryID string) error
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

// entrySummary is the admin-facing view of an entry: its draft data plus
// the computed draft/modified/published status. Field names intentionally
// have no json tags, matching the rest of this API (entities marshal under
// their literal Go field names, e.g. ContentType's ID/Slug/Kind).
type entrySummary struct {
	EntryID       string
	ContentTypeID string
	Data          map[string]any
	Status        string
	Locale        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CreatedBy     string
	UpdatedBy     string
}

func toSummary(doc *entity.Document, status string) entrySummary {
	return entrySummary{
		EntryID:       doc.EntryID,
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

// List returns every entry of a content type with its computed status —
// the admin list view (collection-type panels).
func (h *DocumentHandler) List(w http.ResponseWriter, r *http.Request) {
	contentTypeID := r.URL.Query().Get("contentType")
	drafts, err := h.uc.GetAll(r.Context(), contentTypeID)
	if err != nil {
		writeErr(w, err)
		return
	}
	summaries := make([]entrySummary, 0, len(drafts))
	for _, draft := range drafts {
		_, status, err := h.uc.GetForEdit(r.Context(), draft.EntryID)
		if err != nil {
			writeErr(w, err)
			return
		}
		summaries = append(summaries, toSummary(draft, status))
	}
	writeJSON(w, http.StatusOK, summaries)
}

// Create saves a brand-new entry's draft (collection-type "add entry").
func (h *DocumentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req documentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	doc := &entity.Document{ContentTypeID: req.ContentTypeID, Data: req.Data}
	saved, err := h.uc.Save(r.Context(), doc, middleware.UserID(r.Context()))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toSummary(saved, "draft"))
}

// GetByID returns the draft + computed status for the admin edit screen.
func (h *DocumentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	draft, status, err := h.uc.GetForEdit(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toSummary(draft, status))
}

// GetPublic resolves an entry's published record only — the public/content
// read path. Returns 404 if the entry has never been published, however
// recent its draft.
func (h *DocumentHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	doc, err := h.uc.GetPublished(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

// Update saves changes to an existing entry's draft.
func (h *DocumentHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req documentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	doc := &entity.Document{
		EntryID:       r.PathValue("id"),
		ContentTypeID: req.ContentTypeID,
		Data:          req.Data,
	}
	saved, err := h.uc.Save(r.Context(), doc, middleware.UserID(r.Context()))
	if err != nil {
		writeErr(w, err)
		return
	}
	_, status, err := h.uc.GetForEdit(r.Context(), saved.EntryID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toSummary(saved, status))
}

func (h *DocumentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *DocumentHandler) Publish(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Publish(r.Context(), r.PathValue("id"), middleware.UserID(r.Context())); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "published"})
}

func (h *DocumentHandler) Unpublish(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Unpublish(r.Context(), r.PathValue("id")); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "draft"})
}
