package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/middleware"
	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type documentUseCase interface {
	Save(ctx context.Context, contentTypeSlug string, doc *entity.Document, fields []entity.FieldDefinition, userID string) (*entity.Document, error)
	GetForEdit(ctx context.Context, contentTypeSlug, documentID, locale string, fields []entity.FieldDefinition) (*entity.Document, string, error)
	GetPublished(ctx context.Context, contentTypeSlug, documentID, locale string, fields []entity.FieldDefinition) (*entity.Document, error)
	Publish(ctx context.Context, contentTypeSlug, documentID, locale string, fields []entity.FieldDefinition, userID string) error
	Unpublish(ctx context.Context, contentTypeSlug, documentID, locale string, fields []entity.FieldDefinition) error
	Delete(ctx context.Context, contentTypeSlug, documentID string, fields []entity.FieldDefinition) error
	Duplicate(ctx context.Context, contentTypeSlug, sourceDocumentID, locale string, fields []entity.FieldDefinition, userID string) (*entity.Document, error)

	GetSingleType(ctx context.Context, contentTypeSlug, locale string, fields []entity.FieldDefinition) (*entity.Document, string, error)
	SaveSingleType(ctx context.Context, contentTypeSlug string, data map[string]any, locale string, fields []entity.FieldDefinition, userID string) (*entity.Document, error)
	PublishSingleType(ctx context.Context, contentTypeSlug, locale string, fields []entity.FieldDefinition, userID string) error
	UnpublishSingleType(ctx context.Context, contentTypeSlug, locale string, fields []entity.FieldDefinition) error
	GetAllPaginated(ctx context.Context, contentTypeSlug string, start, size int, locale string, fields []entity.FieldDefinition, orderBy string, sortDir int, filters []entity.FilterNode) ([]*entity.Document, []string, int64, error)
}

type documentContentTypeUseCase interface {
	FindBySlug(ctx context.Context, slug string) (*entity.ContentType, error)
}

type userDisplayNameResolver interface {
	FindByIDs(ctx context.Context, ids []string) ([]*entity.User, error)
}

type DocumentHandler struct {
	usecase      documentUseCase
	contentType  documentContentTypeUseCase
	userResolver userDisplayNameResolver
}

func NewDocumentHandler(usecase documentUseCase, contentType documentContentTypeUseCase, userResolver userDisplayNameResolver) *DocumentHandler {
	return &DocumentHandler{usecase: usecase, contentType: contentType, userResolver: userResolver}
}

var allowedOrderBy = map[string]bool{
	"id":        true,
	"createdAt": true,
	"updatedAt": true,
}

func ginSortParams(c *gin.Context) (orderBy string, sortDir int, ok bool) {
	orderBy = c.DefaultQuery("orderBy", "id")
	if !allowedOrderBy[orderBy] {
		ginWriteError(c, http.StatusBadRequest, "invalid orderBy; allowed: id, createdAt, updatedAt")
		return "", 0, false
	}
	sortDirStr := strings.ToLower(c.DefaultQuery("sortDir", "desc"))
	switch sortDirStr {
	case "asc":
		sortDir = 1
	case "desc":
		sortDir = -1
	default:
		ginWriteError(c, http.StatusBadRequest, "invalid sortDir; allowed: asc, desc")
		return "", 0, false
	}
	return orderBy, sortDir, true
}

type documentRequest struct {
	Data map[string]any `json:"data"`
}

type documentResponse struct {
	Data   map[string]any `json:"data"`
	Status string         `json:"status"`
}

type publicDocumentResponse struct {
	Data map[string]any `json:"data"`
}

type paginatedListItem struct {
	Data   map[string]any `json:"data"`
	Status string         `json:"status"`
}

type paginatedResponse struct {
	Items      []paginatedListItem `json:"items"`
	Total      int64               `json:"total"`
	Start      int                 `json:"start"`
	Size       int                 `json:"size"`
	ListFields []string            `json:"listFields,omitempty"`
}

func mergeDocData(doc *entity.Document) map[string]any {
	merged := make(map[string]any, len(doc.Fields)+7)
	for k, v := range doc.Fields {
		merged[k] = v
	}
	merged["id"] = doc.GormID
	merged["documentId"] = doc.DocumentID
	merged["locale"] = doc.Locale
	merged["createdAt"] = doc.CreatedAt
	merged["updatedAt"] = doc.UpdatedAt
	merged["createdBy"] = doc.CreatedBy
	merged["updatedBy"] = doc.UpdatedBy
	return merged
}

func toDocResponse(doc *entity.Document, status string) documentResponse {
	return documentResponse{
		Data:   mergeDocData(doc),
		Status: status,
	}
}

func mergeListItemData(doc *entity.Document, projectedData map[string]any) map[string]any {
	merged := make(map[string]any, len(projectedData)+7)
	for k, v := range projectedData {
		merged[k] = v
	}
	merged["id"] = doc.GormID
	merged["documentId"] = doc.DocumentID
	merged["locale"] = doc.Locale
	merged["createdAt"] = doc.CreatedAt
	merged["updatedAt"] = doc.UpdatedAt
	merged["createdBy"] = doc.CreatedBy
	merged["updatedBy"] = doc.UpdatedBy
	return merged
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

func (h *DocumentHandler) resolveFields(ginCtx *gin.Context, slug string) []entity.FieldDefinition {
	ct, err := h.contentType.FindBySlug(ginCtx.Request.Context(), slug)
	if err != nil {
		return nil
	}
	return ct.Fields
}

// --- Single-type handlers ---

func (h *DocumentHandler) GetSingleType(ginCtx *gin.Context) {
	slug := ginCtx.Param("slug")
	fields := h.resolveFields(ginCtx, slug)
	doc, status, err := h.usecase.GetSingleType(ginCtx.Request.Context(), slug, ginCtx.Query("locale"), fields)
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusOK, toDocResponse(doc, status))
}

func (h *DocumentHandler) SaveSingleType(ginCtx *gin.Context) {
	var req documentRequest
	if err := ginCtx.ShouldBindJSON(&req); err != nil {
		ginWriteError(ginCtx, http.StatusBadRequest, "invalid request body")
		return
	}
	slug := ginCtx.Param("slug")
	fields := h.resolveFields(ginCtx, slug)
	saved, err := h.usecase.SaveSingleType(ginCtx.Request.Context(), slug, req.Data, ginCtx.Query("locale"), fields, middleware.UserID(ginCtx.Request.Context()))
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	doc, status, err := h.usecase.GetSingleType(ginCtx.Request.Context(), slug, saved.Locale, fields)
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusOK, toDocResponse(doc, status))
}

func (h *DocumentHandler) PublishSingleType(ginCtx *gin.Context) {
	slug := ginCtx.Param("slug")
	fields := h.resolveFields(ginCtx, slug)
	if err := h.usecase.PublishSingleType(ginCtx.Request.Context(), slug, ginCtx.Query("locale"), fields, middleware.UserID(ginCtx.Request.Context())); err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusOK, gin.H{"status": "published"})
}

func (h *DocumentHandler) UnpublishSingleType(ginCtx *gin.Context) {
	slug := ginCtx.Param("slug")
	fields := h.resolveFields(ginCtx, slug)
	if err := h.usecase.UnpublishSingleType(ginCtx.Request.Context(), slug, ginCtx.Query("locale"), fields); err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusOK, gin.H{"status": "draft"})
}

func (h *DocumentHandler) resolveUserDisplayNames(ctx context.Context, userIDs []string) map[string]string {
	nameMap := make(map[string]string, len(userIDs))
	if len(userIDs) == 0 || h.userResolver == nil {
		return nameMap
	}
	unique := make(map[string]bool, len(userIDs))
	var deduped []string
	for _, id := range userIDs {
		if id != "" && !unique[id] {
			unique[id] = true
			deduped = append(deduped, id)
		}
	}
	users, err := h.userResolver.FindByIDs(ctx, deduped)
	if err != nil {
		return nameMap
	}
	for _, user := range users {
		nameMap[user.DocumentID] = user.DisplayName
	}
	return nameMap
}

// --- Collection-type handlers ---

func (h *DocumentHandler) ListCollection(ginCtx *gin.Context) {
	slug := ginCtx.Param("slug")
	start, size := ginPaginationParams(ginCtx)

	orderBy, sortDir, ok := ginSortParams(ginCtx)
	if !ok {
		return
	}

	ct, err := h.contentType.FindBySlug(ginCtx.Request.Context(), slug)
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	listFields := ct.ListFields
	if len(listFields) == 0 && len(ct.Fields) > 0 {
		var flat []entity.FieldDefinition
		for _, field := range ct.Fields {
			if field.Type != "component" {
				flat = append(flat, field)
			}
		}
		limit := 3
		if len(flat) < limit {
			limit = len(flat)
		}
		listFields = make([]string, limit)
		for idx := 0; idx < limit; idx++ {
			listFields[idx] = flat[idx].Name
		}
	}

	docs, statuses, total, err := h.usecase.GetAllPaginated(ginCtx.Request.Context(), slug, start, size, ginCtx.Query("locale"), ct.Fields, orderBy, sortDir, nil)
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}

	updatedByIDs := make([]string, len(docs))
	for i, doc := range docs {
		updatedByIDs[i] = doc.UpdatedBy
	}
	nameMap := h.resolveUserDisplayNames(ginCtx.Request.Context(), updatedByIDs)

	items := make([]paginatedListItem, len(docs))
	for i, doc := range docs {
		data := mergeListItemData(doc, projectData(doc.Fields, listFields))
		if name, ok := nameMap[doc.UpdatedBy]; ok {
			data["updatedByName"] = name
		} else {
			data["updatedByName"] = doc.UpdatedBy
		}
		items[i] = paginatedListItem{
			Data:   data,
			Status: statuses[i],
		}
	}
	ginCtx.JSON(http.StatusOK, paginatedResponse{
		Items:      items,
		Total:      total,
		Start:      start,
		Size:       size,
		ListFields: ct.ListFields,
	})
}

func (h *DocumentHandler) GetCollection(ginCtx *gin.Context) {
	slug := ginCtx.Param("slug")
	documentID := ginCtx.Param("documentId")
	fields := h.resolveFields(ginCtx, slug)
	draft, status, err := h.usecase.GetForEdit(ginCtx.Request.Context(), slug, documentID, ginCtx.Query("locale"), fields)
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	resp := toDocResponse(draft, status)
	nameMap := h.resolveUserDisplayNames(ginCtx.Request.Context(), []string{draft.UpdatedBy})
	if name, ok := nameMap[draft.UpdatedBy]; ok {
		resp.Data["updatedByName"] = name
	} else {
		resp.Data["updatedByName"] = draft.UpdatedBy
	}
	ginCtx.JSON(http.StatusOK, resp)
}

func (h *DocumentHandler) CreateCollection(ginCtx *gin.Context) {
	var req documentRequest
	if err := ginCtx.ShouldBindJSON(&req); err != nil {
		ginWriteError(ginCtx, http.StatusBadRequest, "invalid request body")
		return
	}
	slug := ginCtx.Param("slug")
	fields := h.resolveFields(ginCtx, slug)
	doc := &entity.Document{Fields: req.Data, Locale: ginCtx.Query("locale")}
	saved, err := h.usecase.Save(ginCtx.Request.Context(), slug, doc, fields, middleware.UserID(ginCtx.Request.Context()))
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusCreated, toDocResponse(saved, "draft"))
}

func (h *DocumentHandler) UpdateCollection(ginCtx *gin.Context) {
	var req documentRequest
	if err := ginCtx.ShouldBindJSON(&req); err != nil {
		ginWriteError(ginCtx, http.StatusBadRequest, "invalid request body")
		return
	}
	slug := ginCtx.Param("slug")
	fields := h.resolveFields(ginCtx, slug)
	documentID := ginCtx.Param("documentId")
	doc := &entity.Document{
		DocumentID: documentID,
		Fields:     req.Data,
		Locale:     ginCtx.Query("locale"),
	}
	saved, err := h.usecase.Save(ginCtx.Request.Context(), slug, doc, fields, middleware.UserID(ginCtx.Request.Context()))
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	_, status, err := h.usecase.GetForEdit(ginCtx.Request.Context(), slug, saved.DocumentID, saved.Locale, fields)
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusOK, toDocResponse(saved, status))
}

func (h *DocumentHandler) DeleteCollection(ginCtx *gin.Context) {
	slug := ginCtx.Param("slug")
	documentID := ginCtx.Param("documentId")
	fields := h.resolveFields(ginCtx, slug)
	if err := h.usecase.Delete(ginCtx.Request.Context(), slug, documentID, fields); err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.Status(http.StatusNoContent)
}

func (h *DocumentHandler) PublishCollection(ginCtx *gin.Context) {
	slug := ginCtx.Param("slug")
	documentID := ginCtx.Param("documentId")
	fields := h.resolveFields(ginCtx, slug)
	if err := h.usecase.Publish(ginCtx.Request.Context(), slug, documentID, ginCtx.Query("locale"), fields, middleware.UserID(ginCtx.Request.Context())); err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusOK, gin.H{"status": "published"})
}

func (h *DocumentHandler) UnpublishCollection(ginCtx *gin.Context) {
	slug := ginCtx.Param("slug")
	documentID := ginCtx.Param("documentId")
	fields := h.resolveFields(ginCtx, slug)
	if err := h.usecase.Unpublish(ginCtx.Request.Context(), slug, documentID, ginCtx.Query("locale"), fields); err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusOK, gin.H{"status": "draft"})
}

func (h *DocumentHandler) DuplicateCollection(ginCtx *gin.Context) {
	slug := ginCtx.Param("slug")
	documentID := ginCtx.Param("documentId")
	fields := h.resolveFields(ginCtx, slug)
	saved, err := h.usecase.Duplicate(ginCtx.Request.Context(), slug, documentID, ginCtx.Query("locale"), fields, middleware.UserID(ginCtx.Request.Context()))
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusCreated, toDocResponse(saved, "draft"))
}

// --- Public handler ---

func (h *DocumentHandler) GetPublic(ginCtx *gin.Context) {
	slug := ginCtx.Param("slug")
	documentID := ginCtx.Param("documentId")
	fields := h.resolveFields(ginCtx, slug)
	doc, err := h.usecase.GetPublished(ginCtx.Request.Context(), slug, documentID, ginCtx.Query("locale"), fields)
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusOK, publicDocumentResponse{Data: mergeDocData(doc)})
}
