package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/middleware"
	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type documentUseCase interface {
	Save(ctx context.Context, contentTypeSlug string, doc *entity.Document, userID string) (*entity.Document, error)
	GetForEdit(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, string, error)
	GetPublished(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error)
	Publish(ctx context.Context, contentTypeSlug, documentID, locale, userID string) error
	Unpublish(ctx context.Context, contentTypeSlug, documentID, locale string) error
	Delete(ctx context.Context, contentTypeSlug, documentID string) error

	GetSingleType(ctx context.Context, contentTypeSlug, locale string) (*entity.Document, string, error)
	SaveSingleType(ctx context.Context, contentTypeSlug string, data map[string]any, locale, userID string) (*entity.Document, error)
	PublishSingleType(ctx context.Context, contentTypeSlug, locale, userID string) error
	UnpublishSingleType(ctx context.Context, contentTypeSlug, locale string) error
	GetAllPaginated(ctx context.Context, contentTypeSlug string, start, size int, locale string) ([]*entity.Document, []string, int64, error)
}

type documentContentTypeUseCase interface {
	FindBySlug(ctx context.Context, slug string) (*entity.ContentType, error)
}

type DocumentHandler struct {
	uc   documentUseCase
	ctUC documentContentTypeUseCase
}

func NewDocumentHandler(uc documentUseCase, ctUC documentContentTypeUseCase) *DocumentHandler {
	return &DocumentHandler{uc: uc, ctUC: ctUC}
}

type documentRequest struct {
	Data map[string]any `json:"data"`
}

type entrySummary struct {
	DocumentID    string         `json:"documentId"`
	ContentTypeID string         `json:"contentTypeId"`
	Data          map[string]any `json:"data"`
	Status        string         `json:"status"`
	Locale        string         `json:"locale"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	CreatedBy     string         `json:"createdBy"`
	UpdatedBy     string         `json:"updatedBy"`
}

type paginatedListItem struct {
	DocumentID string         `json:"documentId"`
	Data       map[string]any `json:"data"`
	Status     string         `json:"status"`
	Locale     string         `json:"locale"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
}

type paginatedResponse struct {
	Items []paginatedListItem `json:"items"`
	Total int64               `json:"total"`
	Start int                 `json:"start"`
	Size  int                 `json:"size"`
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

func projectData(data map[string]any, fields []string) map[string]any {
	projected := make(map[string]any, len(fields))
	for _, f := range fields {
		if v, ok := data[f]; ok {
			projected[f] = v
		}
	}
	return projected
}

func paginationParams(r *http.Request) (start, size int) {
	start, _ = strconv.Atoi(r.URL.Query().Get("start"))
	size, _ = strconv.Atoi(r.URL.Query().Get("size"))
	if start < 0 {
		start = 0
	}
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	return start, size
}

// --- Single-type handlers ---

func (h *DocumentHandler) GetSingleType(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	doc, status, err := h.uc.GetSingleType(r.Context(), slug, localeParam(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toSummary(doc, status))
}

func (h *DocumentHandler) SaveSingleType(w http.ResponseWriter, r *http.Request) {
	var req documentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	slug := r.PathValue("slug")
	saved, err := h.uc.SaveSingleType(r.Context(), slug, req.Data, localeParam(r), middleware.UserID(r.Context()))
	if err != nil {
		writeErr(w, err)
		return
	}
	doc, status, err := h.uc.GetSingleType(r.Context(), slug, saved.Locale)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toSummary(doc, status))
}

func (h *DocumentHandler) PublishSingleType(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if err := h.uc.PublishSingleType(r.Context(), slug, localeParam(r), middleware.UserID(r.Context())); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "published"})
}

func (h *DocumentHandler) UnpublishSingleType(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if err := h.uc.UnpublishSingleType(r.Context(), slug, localeParam(r)); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "draft"})
}

// --- Collection-type handlers ---

func (h *DocumentHandler) ListCollection(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	start, size := paginationParams(r)

	ct, err := h.ctUC.FindBySlug(r.Context(), slug)
	if err != nil {
		writeErr(w, err)
		return
	}
	listFields := ct.ListFields
	if len(listFields) == 0 && len(ct.Fields) > 0 {
		limit := 3
		if len(ct.Fields) < limit {
			limit = len(ct.Fields)
		}
		listFields = make([]string, limit)
		for i := 0; i < limit; i++ {
			listFields[i] = ct.Fields[i].Name
		}
	}

	docs, statuses, total, err := h.uc.GetAllPaginated(r.Context(), slug, start, size, localeParam(r))
	if err != nil {
		writeErr(w, err)
		return
	}

	items := make([]paginatedListItem, len(docs))
	for i, doc := range docs {
		items[i] = paginatedListItem{
			DocumentID: doc.DocumentID,
			Data:       projectData(doc.Data, listFields),
			Status:     statuses[i],
			Locale:     doc.Locale,
			CreatedAt:  doc.CreatedAt,
			UpdatedAt:  doc.UpdatedAt,
		}
	}
	writeJSON(w, http.StatusOK, paginatedResponse{
		Items: items,
		Total: total,
		Start: start,
		Size:  size,
	})
}

func (h *DocumentHandler) GetCollection(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	documentID := r.PathValue("documentId")
	draft, status, err := h.uc.GetForEdit(r.Context(), slug, documentID, localeParam(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toSummary(draft, status))
}

func (h *DocumentHandler) CreateCollection(w http.ResponseWriter, r *http.Request) {
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

func (h *DocumentHandler) UpdateCollection(w http.ResponseWriter, r *http.Request) {
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

func (h *DocumentHandler) DeleteCollection(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	documentID := r.PathValue("documentId")
	if err := h.uc.Delete(r.Context(), slug, documentID); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *DocumentHandler) PublishCollection(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	documentID := r.PathValue("documentId")
	if err := h.uc.Publish(r.Context(), slug, documentID, localeParam(r), middleware.UserID(r.Context())); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "published"})
}

func (h *DocumentHandler) UnpublishCollection(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	documentID := r.PathValue("documentId")
	if err := h.uc.Unpublish(r.Context(), slug, documentID, localeParam(r)); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "draft"})
}

// --- Public handler (unchanged) ---

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
