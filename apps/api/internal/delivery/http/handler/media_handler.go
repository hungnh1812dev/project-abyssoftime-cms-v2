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
	Upload(ctx context.Context, file io.Reader, filename string) (*entity.MediaAsset, error)
	List(ctx context.Context, page, limit int) ([]*entity.MediaAsset, int64, error)
	Delete(ctx context.Context, id string) error
}

type MediaHandler struct {
	usecase mediaUseCase
}

func NewMediaHandler(usecase mediaUseCase) *MediaHandler {
	return &MediaHandler{usecase: usecase}
}

func (h *MediaHandler) List(ginCtx *gin.Context) {
	page, _ := strconv.Atoi(ginCtx.Query("page"))
	limit, _ := strconv.Atoi(ginCtx.Query("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	items, total, err := h.usecase.List(ginCtx.Request.Context(), page, limit)
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *MediaHandler) Delete(ginCtx *gin.Context) {
	id := ginCtx.Param("id")
	if err := h.usecase.Delete(ginCtx.Request.Context(), id); err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.Status(http.StatusNoContent)
}

func (h *MediaHandler) Upload(ginCtx *gin.Context) {
	fileHeader, err := ginCtx.FormFile("file")
	if err != nil {
		ginWriteError(ginCtx, http.StatusBadRequest, "file field is required")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		ginWriteError(ginCtx, http.StatusBadRequest, "failed to read uploaded file")
		return
	}
	defer file.Close()

	asset, err := h.usecase.Upload(ginCtx.Request.Context(), file, fileHeader.Filename)
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusCreated, asset)
}
