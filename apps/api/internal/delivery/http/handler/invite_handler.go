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
	uc inviteUseCase
}

func NewInviteHandler(uc inviteUseCase) *InviteHandler {
	return &InviteHandler{uc: uc}
}

func (h *InviteHandler) Create(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ginWriteError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := c.GetString("userID")
	inv, plaintext, err := h.uc.Create(c.Request.Context(), actorID, req.Email, entity.Role(req.Role))
	if err != nil {
		ginWriteErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        inv.ID,
		"email":     inv.Email,
		"role":      inv.Role,
		"expiresAt": inv.ExpiresAt,
		"token":     plaintext,
	})
}

func (h *InviteHandler) List(c *gin.Context) {
	invites, err := h.uc.List(c.Request.Context())
	if err != nil {
		ginWriteErr(c, err)
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
	for i, inv := range invites {
		items[i] = inviteItem{
			ID:        inv.ID,
			Email:     inv.Email,
			Role:      string(inv.Role),
			ExpiresAt: inv.ExpiresAt.Format("2006-01-02T15:04:05Z"),
			CreatedBy: inv.CreatedBy,
			CreatedAt: inv.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *InviteHandler) Revoke(c *gin.Context) {
	if err := h.uc.Revoke(c.Request.Context(), c.Param("id")); err != nil {
		ginWriteErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *InviteHandler) Accept(c *gin.Context) {
	var req struct {
		Password    string `json:"password"`
		DisplayName string `json:"displayName"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ginWriteError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	token := c.Param("token")
	user, err := h.uc.Accept(c.Request.Context(), token, req.Password, req.DisplayName)
	if err != nil {
		ginWriteErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          user.ID,
		"email":       user.Email,
		"displayName": user.DisplayName,
		"role":        user.Role,
	})
}
