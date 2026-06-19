package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type accessTokenUseCase interface {
	Create(ctx context.Context, name string, scopes []string, expiresAt *time.Time, createdBy string) (*entity.AccessToken, string, error)
	List(ctx context.Context, page, limit int) ([]*entity.AccessToken, int64, error)
	Delete(ctx context.Context, id string) error
}

type AccessTokenHandler struct {
	uc accessTokenUseCase
}

func NewAccessTokenHandler(uc accessTokenUseCase) *AccessTokenHandler {
	return &AccessTokenHandler{uc: uc}
}

func (h *AccessTokenHandler) Create(c *gin.Context) {
	var req struct {
		Name      string   `json:"name"`
		Scopes    []string `json:"scopes"`
		ExpiresIn *string  `json:"expiresIn"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		ginWriteError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	var expiresAt *time.Time
	if req.ExpiresIn != nil {
		d, err := time.ParseDuration(*req.ExpiresIn)
		if err == nil {
			t := time.Now().UTC().Add(d)
			expiresAt = &t
		}
	}

	createdBy := c.GetString("userID")
	token, plaintext, err := h.uc.Create(c.Request.Context(), req.Name, req.Scopes, expiresAt, createdBy)
	if err != nil {
		ginWriteErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        token.ID,
		"name":      token.Name,
		"prefix":    token.Prefix,
		"scopes":    token.Scopes,
		"expiresAt": token.ExpiresAt,
		"createdAt": token.CreatedAt,
		"token":     plaintext,
	})
}

func (h *AccessTokenHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	tokens, total, err := h.uc.List(c.Request.Context(), page, limit)
	if err != nil {
		ginWriteErr(c, err)
		return
	}

	type tokenItem struct {
		ID         string     `json:"id"`
		Name       string     `json:"name"`
		Prefix     string     `json:"prefix"`
		Scopes     []string   `json:"scopes"`
		ExpiresAt  *time.Time `json:"expiresAt"`
		LastUsedAt *time.Time `json:"lastUsedAt"`
		CreatedAt  time.Time  `json:"createdAt"`
	}
	items := make([]tokenItem, len(tokens))
	for i, t := range tokens {
		items[i] = tokenItem{
			ID:         t.ID,
			Name:       t.Name,
			Prefix:     t.Prefix,
			Scopes:     t.Scopes,
			ExpiresAt:  t.ExpiresAt,
			LastUsedAt: t.LastUsedAt,
			CreatedAt:  t.CreatedAt,
		}
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "limit": limit})
}

func (h *AccessTokenHandler) Delete(c *gin.Context) {
	if err := h.uc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		ginWriteErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
