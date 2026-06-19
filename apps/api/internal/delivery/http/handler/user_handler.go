package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type userUseCase interface {
	List(ctx context.Context, page, limit int) ([]*entity.User, int64, error)
	GetByID(ctx context.Context, id string) (*entity.User, error)
	UpdateRole(ctx context.Context, actorID, targetID string, newRole entity.Role) error
	Delete(ctx context.Context, actorID, targetID string) error
}

type UserHandler struct {
	uc userUseCase
}

func NewUserHandler(uc userUseCase) *UserHandler {
	return &UserHandler{uc: uc}
}

type userResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func toUserResponse(u *entity.User) userResponse {
	return userResponse{ID: u.ID, Email: u.Email, Role: string(u.Role)}
}

func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	users, total, err := h.uc.List(c.Request.Context(), page, limit)
	if err != nil {
		ginWriteErr(c, err)
		return
	}

	items := make([]userResponse, len(users))
	for i, u := range users {
		items[i] = toUserResponse(u)
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "limit": limit})
}

func (h *UserHandler) Get(c *gin.Context) {
	user, err := h.uc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		ginWriteErr(c, err)
		return
	}
	c.JSON(http.StatusOK, toUserResponse(user))
}

func (h *UserHandler) UpdateRole(c *gin.Context) {
	var req struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ginWriteError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := c.GetString("userID")
	targetID := c.Param("id")

	if err := h.uc.UpdateRole(c.Request.Context(), actorID, targetID, entity.Role(req.Role)); err != nil {
		ginWriteErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *UserHandler) Delete(c *gin.Context) {
	actorID := c.GetString("userID")
	targetID := c.Param("id")

	if err := h.uc.Delete(c.Request.Context(), actorID, targetID); err != nil {
		ginWriteErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
