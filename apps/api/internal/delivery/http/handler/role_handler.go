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
	uc    roleUseCase
	cache *middleware.RoleCache
}

func NewRoleHandler(uc roleUseCase, cache *middleware.RoleCache) *RoleHandler {
	return &RoleHandler{uc: uc, cache: cache}
}

func (h *RoleHandler) List(c *gin.Context) {
	roles, err := h.uc.FindAll(c.Request.Context())
	if err != nil {
		ginWriteErr(c, err)
		return
	}
	c.JSON(http.StatusOK, roles)
}

func (h *RoleHandler) Get(c *gin.Context) {
	r, err := h.uc.FindByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		ginWriteErr(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req struct {
		Name        string   `json:"name"`
		Slug        string   `json:"slug"`
		Permissions []string `json:"permissions"`
		Level       int      `json:"level"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ginWriteError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	callerLevel := h.cache.GetLevel(c.GetString("role"))

	created, err := h.uc.Create(c.Request.Context(), role.CreateRoleInput{
		Name:        req.Name,
		Slug:        req.Slug,
		Permissions: req.Permissions,
		Level:       req.Level,
	}, callerLevel)
	if err != nil {
		ginWriteErr(c, err)
		return
	}

	h.reloadCache(c.Request.Context())
	c.JSON(http.StatusCreated, created)
}

func (h *RoleHandler) Update(c *gin.Context) {
	var req struct {
		Name        *string   `json:"name"`
		Permissions *[]string `json:"permissions"`
		Level       *int      `json:"level"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ginWriteError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	callerLevel := h.cache.GetLevel(c.GetString("role"))

	updated, err := h.uc.Update(c.Request.Context(), c.Param("id"), role.UpdateRoleInput{
		Name:        req.Name,
		Permissions: req.Permissions,
		Level:       req.Level,
	}, callerLevel)
	if err != nil {
		ginWriteErr(c, err)
		return
	}

	h.reloadCache(c.Request.Context())
	c.JSON(http.StatusOK, updated)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	if err := h.uc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		ginWriteErr(c, err)
		return
	}

	h.reloadCache(c.Request.Context())
	c.Status(http.StatusNoContent)
}

func (h *RoleHandler) reloadCache(ctx context.Context) {
	roles, err := h.uc.FindAll(ctx)
	if err == nil {
		h.cache.Load(roles)
	}
}
