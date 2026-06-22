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
	UpdateRole(ctx context.Context, actorID, targetID, newRoleID string) error
	Delete(ctx context.Context, actorID, targetID string) error
}

type UserHandler struct {
	usecase userUseCase
}

func NewUserHandler(usecase userUseCase) *UserHandler {
	return &UserHandler{usecase: usecase}
}

type userResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
	RoleID      string `json:"roleId"`
}

func toUserResponse(user *entity.User) userResponse {
	return userResponse{ID: user.DocumentID, Email: user.Email, DisplayName: user.DisplayName, Role: string(user.Role), RoleID: user.RoleID}
}

func (h *UserHandler) List(ginCtx *gin.Context) {
	page, _ := strconv.Atoi(ginCtx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ginCtx.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	users, total, err := h.usecase.List(ginCtx.Request.Context(), page, limit)
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}

	items := make([]userResponse, len(users))
	for i, user := range users {
		items[i] = toUserResponse(user)
	}
	ginCtx.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "limit": limit})
}

func (h *UserHandler) Get(ginCtx *gin.Context) {
	user, err := h.usecase.GetByID(ginCtx.Request.Context(), ginCtx.Param("id"))
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusOK, toUserResponse(user))
}

func (h *UserHandler) UpdateRole(ginCtx *gin.Context) {
	var req struct {
		RoleID string `json:"roleId"`
	}
	if err := ginCtx.ShouldBindJSON(&req); err != nil || req.RoleID == "" {
		ginWriteError(ginCtx, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := ginCtx.GetString("userID")
	targetID := ginCtx.Param("id")

	if err := h.usecase.UpdateRole(ginCtx.Request.Context(), actorID, targetID, req.RoleID); err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.Status(http.StatusNoContent)
}

func (h *UserHandler) Delete(ginCtx *gin.Context) {
	actorID := ginCtx.GetString("userID")
	targetID := ginCtx.Param("id")

	if err := h.usecase.Delete(ginCtx.Request.Context(), actorID, targetID); err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.Status(http.StatusNoContent)
}
