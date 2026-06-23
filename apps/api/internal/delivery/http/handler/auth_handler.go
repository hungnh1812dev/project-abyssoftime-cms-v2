package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

const RefreshCookieName = "refresh_token"

const (
	refreshCookieMaxAgeDefault  = 7 * 24 * 60 * 60  // 7 days
	refreshCookieMaxAgeRemember = 30 * 24 * 60 * 60 // 30 days
)

func noStoreHeader(ginCtx *gin.Context) {
	ginCtx.Header("Cache-Control", "no-store")
}

type authUseCase interface {
	Register(ctx context.Context, email, password, displayName string) (*entity.User, error)
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, userID string) error
	SetupStatus(ctx context.Context) (bool, error)
}

type AuthHandler struct {
	usecase        authUseCase
	cookieSecure   bool
	cookieSameSite http.SameSite
}

func NewAuthHandler(usecase authUseCase, cookieSecure bool, cookieSameSite http.SameSite) *AuthHandler {
	return &AuthHandler{usecase: usecase, cookieSecure: cookieSecure, cookieSameSite: cookieSameSite}
}

func (h *AuthHandler) Register(ginCtx *gin.Context) {
	var req struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"displayName"`
	}
	if err := ginCtx.ShouldBindJSON(&req); err != nil {
		ginWriteError(ginCtx, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.usecase.Register(ginCtx.Request.Context(), req.Email, req.Password, req.DisplayName)
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

func (h *AuthHandler) Login(ginCtx *gin.Context) {
	var req struct {
		Email      string `json:"email"`
		Password   string `json:"password"`
		RememberMe bool   `json:"rememberMe"`
	}
	if err := ginCtx.ShouldBindJSON(&req); err != nil {
		ginWriteError(ginCtx, http.StatusBadRequest, "invalid request body")
		return
	}

	access, refresh, err := h.usecase.Login(ginCtx.Request.Context(), req.Email, req.Password)
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}

	maxAge := refreshCookieMaxAgeDefault
	if req.RememberMe {
		maxAge = refreshCookieMaxAgeRemember
	}
	ginCtx.SetSameSite(h.cookieSameSite)
	ginCtx.SetCookie(RefreshCookieName, refresh, maxAge, "/", "", h.cookieSecure, true)

	noStoreHeader(ginCtx)
	ginCtx.JSON(http.StatusOK, gin.H{
		"accessToken": access,
	})
}

func (h *AuthHandler) Refresh(ginCtx *gin.Context) {
	cookieVal, err := ginCtx.Cookie(RefreshCookieName)
	if err != nil {
		ginWriteError(ginCtx, http.StatusUnauthorized, "missing refresh token")
		return
	}

	access, refresh, err := h.usecase.RefreshToken(ginCtx.Request.Context(), cookieVal)
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}

	ginCtx.SetSameSite(h.cookieSameSite)
	ginCtx.SetCookie(RefreshCookieName, refresh, refreshCookieMaxAgeRemember, "/", "", h.cookieSecure, true)

	noStoreHeader(ginCtx)
	ginCtx.JSON(http.StatusOK, gin.H{
		"accessToken": access,
	})
}

func (h *AuthHandler) SetupStatus(ginCtx *gin.Context) {
	adminExists, err := h.usecase.SetupStatus(ginCtx.Request.Context())
	if err != nil {
		ginWriteErr(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusOK, gin.H{
		"adminExists": adminExists,
	})
}

func (h *AuthHandler) Logout(ginCtx *gin.Context) {
	ginCtx.SetSameSite(h.cookieSameSite)
	ginCtx.SetCookie(RefreshCookieName, "", -1, "/", "", h.cookieSecure, true)

	noStoreHeader(ginCtx)
	ginCtx.JSON(http.StatusOK, gin.H{"ok": true})
}
