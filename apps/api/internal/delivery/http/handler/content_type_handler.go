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
}

type ContentTypeHandler struct {
	uc contentTypeUseCase
}

func NewContentTypeHandler(uc contentTypeUseCase) *ContentTypeHandler {
	return &ContentTypeHandler{uc: uc}
}

type contentTypeSummary struct {
	ID   string             `json:"ID"`
	Name string             `json:"Name"`
	Slug string             `json:"Slug"`
	Kind entity.ContentKind `json:"Kind"`
}

func (h *ContentTypeHandler) ListSummary(c *gin.Context) {
	cts, err := h.uc.FindAll(c.Request.Context())
	if err != nil {
		ginWriteErr(c, err)
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
	c.JSON(http.StatusOK, summaries)
}

func (h *ContentTypeHandler) Get(c *gin.Context) {
	identifier := c.Param("identifier")
	var (
		ct  *entity.ContentType
		err error
	)
	if objectIDRe.MatchString(identifier) {
		ct, err = h.uc.FindByID(c.Request.Context(), identifier)
	} else {
		ct, err = h.uc.FindBySlug(c.Request.Context(), identifier)
	}
	if err != nil {
		ginWriteErr(c, err)
		return
	}
	c.JSON(http.StatusOK, ct)
}
