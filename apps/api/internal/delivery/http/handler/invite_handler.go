package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type inviteUseCase interface {
	Create(ctx context.Context, actorID, email string, role entity.Role) (*entity.Invite, string, error)
	List(ctx context.Context) ([]*entity.Invite, error)
	Revoke(ctx context.Context, id string) error
	Accept(ctx context.Context, token, password, displayName string) (*entity.User, error)
}

type InviteHandler struct {
	usecase inviteUseCase
}

func NewInviteHandler(usecase inviteUseCase) *InviteHandler {
	return &InviteHandler{usecase: usecase}
}

func (h *InviteHandler) Create(ginCtx *gin.Context) {
	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := ginCtx.ShouldBindJSON(&req); err != nil {
		ginWriteError(ginCtx, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := ginCtx.GetString("userID")
	invite, plaintext, err := h.usecase.Create(ginCtx.Request.Context(), actorID, req.Email, entity.Role(req.Role))
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}

	ginCtx.JSON(http.StatusCreated, gin.H{
		"id":        invite.DocumentID,
		"email":     invite.Email,
		"role":      invite.Role,
		"expiresAt": invite.ExpiresAt,
		"token":     plaintext,
	})
}

func (h *InviteHandler) List(ginCtx *gin.Context) {
	invites, err := h.usecase.List(ginCtx.Request.Context())
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}

	type inviteItem struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		Role      string `json:"role"`
		ExpiresAt string `json:"expiresAt"`
		CreatedBy string `json:"createdBy"`
		CreatedAt string `json:"createdAt"`
	}
	items := make([]inviteItem, len(invites))
	for i, invite := range invites {
		items[i] = inviteItem{
			ID:        invite.DocumentID,
			Email:     invite.Email,
			Role:      string(invite.Role),
			ExpiresAt: invite.ExpiresAt.Format("2006-01-02T15:04:05Z"),
			CreatedBy: invite.CreatedBy,
			CreatedAt: invite.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}
	ginCtx.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *InviteHandler) Revoke(ginCtx *gin.Context) {
	if err := h.usecase.Revoke(ginCtx.Request.Context(), ginCtx.Param("id")); err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.Status(http.StatusNoContent)
}

func (h *InviteHandler) Accept(ginCtx *gin.Context) {
	var req struct {
		Password    string `json:"password"`
		DisplayName string `json:"displayName"`
	}
	if err := ginCtx.ShouldBindJSON(&req); err != nil {
		ginWriteError(ginCtx, http.StatusBadRequest, "invalid request body")
		return
	}

	token := ginCtx.Param("token")
	user, err := h.usecase.Accept(ginCtx.Request.Context(), token, req.Password, req.DisplayName)
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}

	ginCtx.JSON(http.StatusCreated, gin.H{
		"id":          user.ID,
		"email":       user.Email,
		"displayName": user.DisplayName,
		"role":        user.Role,
	})
}
