package handler

import (
	"context"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

var objectIDRe = regexp.MustCompile(`^[a-f0-9]{24}$`)

type contentTypeUseCase interface {
	FindByID(ctx context.Context, id string) (*entity.ContentType, error)
	FindBySlug(ctx context.Context, slug string) (*entity.ContentType, error)
	FindAll(ctx context.Context) ([]*entity.ContentType, error)
	Update(ctx context.Context, ct *entity.ContentType) error
}

type ContentTypeHandler struct {
	usecase contentTypeUseCase
}

func NewContentTypeHandler(usecase contentTypeUseCase) *ContentTypeHandler {
	return &ContentTypeHandler{usecase: usecase}
}

type contentTypeSummary struct {
	ID   string             `json:"ID"`
	Name string             `json:"Name"`
	Slug string             `json:"Slug"`
	Kind entity.ContentKind `json:"Kind"`
}

func (h *ContentTypeHandler) ListSummary(ginCtx *gin.Context) {
	cts, err := h.usecase.FindAll(ginCtx.Request.Context())
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	summaries := make([]contentTypeSummary, len(cts))
	for i, ct := range cts {
		summaries[i] = contentTypeSummary{
			ID:   ct.DocumentID,
			Name: ct.Name,
			Slug: ct.Slug,
			Kind: ct.Kind,
		}
	}
	ginCtx.JSON(http.StatusOK, summaries)
}

func (h *ContentTypeHandler) Get(ginCtx *gin.Context) {
	identifier := ginCtx.Param("identifier")
	var (
		contentType *entity.ContentType
		err         error
	)
	if objectIDRe.MatchString(identifier) {
		contentType, err = h.usecase.FindByID(ginCtx.Request.Context(), identifier)
	} else {
		contentType, err = h.usecase.FindBySlug(ginCtx.Request.Context(), identifier)
	}
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusOK, contentType)
}

var knownSystemFields = map[string]bool{
	"createdAt":     true,
	"updatedAt":     true,
	"updatedByName": true,
}

func (h *ContentTypeHandler) UpdateListFields(ginCtx *gin.Context) {
	slug := ginCtx.Param("slug")

	var req struct {
		ListFields []string `json:"listFields"`
	}
	if err := ginCtx.ShouldBindJSON(&req); err != nil {
		ginWriteError(ginCtx, http.StatusBadRequest, "invalid request body")
		return
	}

	contentType, err := h.usecase.FindBySlug(ginCtx.Request.Context(), slug)
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}

	fieldNames := make(map[string]bool, len(contentType.Fields))
	for _, field := range contentType.Fields {
		if field.Type == "layout" {
			for _, child := range field.Fields {
				fieldNames[child.Name] = true
			}
		} else if field.Type != "component" {
			fieldNames[field.Name] = true
		}
	}
	for _, entry := range req.ListFields {
		if !fieldNames[entry] && !knownSystemFields[entry] {
			ginWriteError(ginCtx, http.StatusBadRequest, "invalid field: "+entry)
			return
		}
	}

	contentType.ListFields = req.ListFields
	if err := h.usecase.Update(ginCtx.Request.Context(), contentType); err != nil {
		ginWriteErr(ginCtx, err)
		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"listFields": contentType.ListFields})
}
