package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

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

func ginPaginationParams(c *gin.Context) (start, size int) {
	start, _ = strconv.Atoi(c.Query("start"))
	size, _ = strconv.Atoi(c.Query("size"))
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

func (h *DocumentHandler) GetSingleType(c *gin.Context) {
	slug := c.Param("slug")
	doc, status, err := h.uc.GetSingleType(c.Request.Context(), slug, c.Query("locale"))
	if err != nil {
		ginWriteErr(c, err)
		return
	}
	c.JSON(http.StatusOK, toSummary(doc, status))
}

func (h *DocumentHandler) SaveSingleType(c *gin.Context) {
	var req documentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginWriteError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	slug := c.Param("slug")
	saved, err := h.uc.SaveSingleType(c.Request.Context(), slug, req.Data, c.Query("locale"), middleware.UserID(c.Request.Context()))
	if err != nil {
		ginWriteErr(c, err)
		return
	}
	doc, status, err := h.uc.GetSingleType(c.Request.Context(), slug, saved.Locale)
	if err != nil {
		ginWriteErr(c, err)
		return
	}
	c.JSON(http.StatusOK, toSummary(doc, status))
}

func (h *DocumentHandler) PublishSingleType(c *gin.Context) {
	slug := c.Param("slug")
	if err := h.uc.PublishSingleType(c.Request.Context(), slug, c.Query("locale"), middleware.UserID(c.Request.Context())); err != nil {
		ginWriteErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "published"})
}

func (h *DocumentHandler) UnpublishSingleType(c *gin.Context) {
	slug := c.Param("slug")
	if err := h.uc.UnpublishSingleType(c.Request.Context(), slug, c.Query("locale")); err != nil {
		ginWriteErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "draft"})
}

// --- Collection-type handlers ---

func (h *DocumentHandler) ListCollection(c *gin.Context) {
	slug := c.Param("slug")
	start, size := ginPaginationParams(c)

	ct, err := h.ctUC.FindBySlug(c.Request.Context(), slug)
	if err != nil {
		ginWriteErr(c, err)
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

	docs, statuses, total, err := h.uc.GetAllPaginated(c.Request.Context(), slug, start, size, c.Query("locale"))
	if err != nil {
		ginWriteErr(c, err)
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
	c.JSON(http.StatusOK, paginatedResponse{
		Items: items,
		Total: total,
		Start: start,
		Size:  size,
	})
}

func (h *DocumentHandler) GetCollection(c *gin.Context) {
	slug := c.Param("slug")
	documentID := c.Param("documentId")
	draft, status, err := h.uc.GetForEdit(c.Request.Context(), slug, documentID, c.Query("locale"))
	if err != nil {
		ginWriteErr(c, err)
		return
	}
	c.JSON(http.StatusOK, toSummary(draft, status))
}

func (h *DocumentHandler) CreateCollection(c *gin.Context) {
	var req documentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginWriteError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	slug := c.Param("slug")
	doc := &entity.Document{Data: req.Data, Locale: c.Query("locale")}
	saved, err := h.uc.Save(c.Request.Context(), slug, doc, middleware.UserID(c.Request.Context()))
	if err != nil {
		ginWriteErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, toSummary(saved, "draft"))
}

func (h *DocumentHandler) UpdateCollection(c *gin.Context) {
	var req documentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginWriteError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	slug := c.Param("slug")
	documentID := c.Param("documentId")
	doc := &entity.Document{
		DocumentID: documentID,
		Data:       req.Data,
		Locale:     c.Query("locale"),
	}
	saved, err := h.uc.Save(c.Request.Context(), slug, doc, middleware.UserID(c.Request.Context()))
	if err != nil {
		ginWriteErr(c, err)
		return
	}
	_, status, err := h.uc.GetForEdit(c.Request.Context(), slug, saved.DocumentID, saved.Locale)
	if err != nil {
		ginWriteErr(c, err)
		return
	}
	c.JSON(http.StatusOK, toSummary(saved, status))
}

func (h *DocumentHandler) DeleteCollection(c *gin.Context) {
	slug := c.Param("slug")
	documentID := c.Param("documentId")
	if err := h.uc.Delete(c.Request.Context(), slug, documentID); err != nil {
		ginWriteErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *DocumentHandler) PublishCollection(c *gin.Context) {
	slug := c.Param("slug")
	documentID := c.Param("documentId")
	if err := h.uc.Publish(c.Request.Context(), slug, documentID, c.Query("locale"), middleware.UserID(c.Request.Context())); err != nil {
		ginWriteErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "published"})
}

func (h *DocumentHandler) UnpublishCollection(c *gin.Context) {
	slug := c.Param("slug")
	documentID := c.Param("documentId")
	if err := h.uc.Unpublish(c.Request.Context(), slug, documentID, c.Query("locale")); err != nil {
		ginWriteErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "draft"})
}

// --- Public handler ---

func (h *DocumentHandler) GetPublic(c *gin.Context) {
	slug := c.Param("slug")
	documentID := c.Param("documentId")
	doc, err := h.uc.GetPublished(c.Request.Context(), slug, documentID, c.Query("locale"))
	if err != nil {
		ginWriteErr(c, err)
		return
	}
	c.JSON(http.StatusOK, doc)
}
