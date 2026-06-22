package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	locale "project-abyssoftime-cms-v2/api/internal/usecase/locale"
)

type localeUseCase interface {
	List(ctx context.Context) ([]*entity.Locale, error)
	Create(ctx context.Context, input locale.CreateInput) (*entity.Locale, error)
	Update(ctx context.Context, code string, input locale.UpdateInput) (*entity.Locale, error)
	Delete(ctx context.Context, code string) error
}

type LocaleHandler struct {
	usecase localeUseCase
}

func NewLocaleHandler(usecase localeUseCase) *LocaleHandler {
	return &LocaleHandler{usecase: usecase}
}

func (handler *LocaleHandler) List(ginCtx *gin.Context) {
	locales, err := handler.usecase.List(ginCtx.Request.Context())
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusOK, locales)
}

func (handler *LocaleHandler) Create(ginCtx *gin.Context) {
	var req struct {
		Code      string `json:"code"`
		Name      string `json:"name"`
		IsDefault bool   `json:"isDefault"`
	}
	if err := ginCtx.ShouldBindJSON(&req); err != nil {
		ginWriteError(ginCtx, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := handler.usecase.Create(ginCtx.Request.Context(), locale.CreateInput{
		Code:      req.Code,
		Name:      req.Name,
		IsDefault: req.IsDefault,
	})
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusCreated, created)
}

func (handler *LocaleHandler) Update(ginCtx *gin.Context) {
	var req struct {
		Name      *string `json:"name"`
		IsDefault *bool   `json:"isDefault"`
	}
	if err := ginCtx.ShouldBindJSON(&req); err != nil {
		ginWriteError(ginCtx, http.StatusBadRequest, "invalid request body")
		return
	}

	updated, err := handler.usecase.Update(ginCtx.Request.Context(), ginCtx.Param("code"), locale.UpdateInput{
		Name:      req.Name,
		IsDefault: req.IsDefault,
	})
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusOK, updated)
}

func (handler *LocaleHandler) Delete(ginCtx *gin.Context) {
	if err := handler.usecase.Delete(ginCtx.Request.Context(), ginCtx.Param("code")); err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.Status(http.StatusNoContent)
}
