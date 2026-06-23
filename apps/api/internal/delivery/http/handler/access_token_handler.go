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
	usecase accessTokenUseCase
}

func NewAccessTokenHandler(usecase accessTokenUseCase) *AccessTokenHandler {
	return &AccessTokenHandler{usecase: usecase}
}

func (h *AccessTokenHandler) Create(ginCtx *gin.Context) {
	var req struct {
		Name      string   `json:"name"`
		Scopes    []string `json:"scopes"`
		ExpiresIn *string  `json:"expiresIn"`
	}
	if err := ginCtx.ShouldBindJSON(&req); err != nil || req.Name == "" {
		ginWriteError(ginCtx, http.StatusBadRequest, "invalid request body")
		return
	}

	var expiresAt *time.Time
	if req.ExpiresIn != nil {
		duration, err := time.ParseDuration(*req.ExpiresIn)
		if err == nil {
			t := time.Now().UTC().Add(duration)
			expiresAt = &t
		}
	}

	createdBy := ginCtx.GetString("userID")
	token, plaintext, err := h.usecase.Create(ginCtx.Request.Context(), req.Name, req.Scopes, expiresAt, createdBy)
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}

	ginCtx.JSON(http.StatusCreated, gin.H{
		"id":        token.ID,
		"name":      token.Name,
		"prefix":    token.Prefix,
		"scopes":    token.Scopes,
		"expiresAt": token.ExpiresAt,
		"createdAt": token.CreatedAt,
		"token":     plaintext,
	})
}

func (h *AccessTokenHandler) List(ginCtx *gin.Context) {
	page, _ := strconv.Atoi(ginCtx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ginCtx.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	tokens, total, err := h.usecase.List(ginCtx.Request.Context(), page, limit)
	if err != nil {
		ginWriteErr(ginCtx, err)
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
	for i, token := range tokens {
		items[i] = tokenItem{
			ID:         token.DocumentID,
			Name:       token.Name,
			Prefix:     token.Prefix,
			Scopes:     token.Scopes,
			ExpiresAt:  token.ExpiresAt,
			LastUsedAt: token.LastUsedAt,
			CreatedAt:  token.CreatedAt,
		}
	}
	ginCtx.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "limit": limit})
}

func (h *AccessTokenHandler) Delete(ginCtx *gin.Context) {
	if err := h.usecase.Delete(ginCtx.Request.Context(), ginCtx.Param("id")); err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.Status(http.StatusNoContent)
}
