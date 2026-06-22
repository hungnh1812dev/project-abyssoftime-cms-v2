package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/middleware"
	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	role "project-abyssoftime-cms-v2/api/internal/usecase/role"
)

type roleUseCase interface {
	FindAll(ctx context.Context) ([]*entity.RoleEntity, error)
	FindByID(ctx context.Context, documentID string) (*entity.RoleEntity, error)
	Create(ctx context.Context, input role.CreateRoleInput, callerLevel int) (*entity.RoleEntity, error)
	Update(ctx context.Context, documentID string, input role.UpdateRoleInput, callerLevel int) (*entity.RoleEntity, error)
	Delete(ctx context.Context, documentID string) error
}

type RoleHandler struct {
	usecase roleUseCase
	cache   *middleware.RoleCache
}

func NewRoleHandler(usecase roleUseCase, cache *middleware.RoleCache) *RoleHandler {
	return &RoleHandler{usecase: usecase, cache: cache}
}

func (h *RoleHandler) List(ginCtx *gin.Context) {
	roles, err := h.usecase.FindAll(ginCtx.Request.Context())
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusOK, roles)
}

func (h *RoleHandler) Get(ginCtx *gin.Context) {
	roleEntity, err := h.usecase.FindByID(ginCtx.Request.Context(), ginCtx.Param("id"))
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusOK, roleEntity)
}

func (h *RoleHandler) Create(ginCtx *gin.Context) {
	var req struct {
		Name        string   `json:"name"`
		Slug        string   `json:"slug"`
		Permissions []string `json:"permissions"`
		Level       int      `json:"level"`
	}
	if err := ginCtx.ShouldBindJSON(&req); err != nil {
		ginWriteError(ginCtx, http.StatusBadRequest, "invalid request body")
		return
	}

	callerLevel := h.cache.GetLevel(ginCtx.GetString("role"))

	created, err := h.usecase.Create(ginCtx.Request.Context(), role.CreateRoleInput{
		Name:        req.Name,
		Slug:        req.Slug,
		Permissions: req.Permissions,
		Level:       req.Level,
	}, callerLevel)
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}

	h.reloadCache(ginCtx.Request.Context())
	ginCtx.JSON(http.StatusCreated, created)
}

func (h *RoleHandler) Update(ginCtx *gin.Context) {
	var req struct {
		Name        *string   `json:"name"`
		Permissions *[]string `json:"permissions"`
		Level       *int      `json:"level"`
	}
	if err := ginCtx.ShouldBindJSON(&req); err != nil {
		ginWriteError(ginCtx, http.StatusBadRequest, "invalid request body")
		return
	}

	callerLevel := h.cache.GetLevel(ginCtx.GetString("role"))

	updated, err := h.usecase.Update(ginCtx.Request.Context(), ginCtx.Param("id"), role.UpdateRoleInput{
		Name:        req.Name,
		Permissions: req.Permissions,
		Level:       req.Level,
	}, callerLevel)
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}

	h.reloadCache(ginCtx.Request.Context())
	ginCtx.JSON(http.StatusOK, updated)
}

func (h *RoleHandler) Delete(ginCtx *gin.Context) {
	if err := h.usecase.Delete(ginCtx.Request.Context(), ginCtx.Param("id")); err != nil {
		ginWriteErr(ginCtx, err)
		return
	}

	h.reloadCache(ginCtx.Request.Context())
	ginCtx.Status(http.StatusNoContent)
}

func (h *RoleHandler) reloadCache(ctx context.Context) {
	roles, err := h.usecase.FindAll(ctx)
	if err == nil {
		h.cache.Load(roles)
	}
}
