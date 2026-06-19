package handler

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type mediaUseCase interface {
	Upload(ctx context.Context, file io.Reader, filename, documentRef, contentTypeID string) (*entity.MediaAsset, error)
	List(ctx context.Context, page, limit int) ([]*entity.MediaAsset, int64, error)
	Delete(ctx context.Context, id string) error
}

type MediaHandler struct {
	uc mediaUseCase
}

func NewMediaHandler(uc mediaUseCase) *MediaHandler {
	return &MediaHandler{uc: uc}
}

func (h *MediaHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	items, total, err := h.uc.List(c.Request.Context(), page, limit)
	if err != nil {
		ginWriteErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *MediaHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		ginWriteErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *MediaHandler) Upload(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		ginWriteError(c, http.StatusBadRequest, "file field is required")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		ginWriteError(c, http.StatusBadRequest, "failed to read uploaded file")
		return
	}
	defer file.Close()

	documentRef := c.PostForm("documentRef")
	contentTypeID := c.PostForm("contentTypeId")

	asset, err := h.uc.Upload(c.Request.Context(), file, fileHeader.Filename, documentRef, contentTypeID)
	if err != nil {
		ginWriteErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, asset)
}
