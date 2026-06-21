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

type authUseCase interface {
	Register(ctx context.Context, email, password, displayName string) (*entity.User, error)
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, userID string) error
	SetupStatus(ctx context.Context) (bool, error)
}

type AuthHandler struct {
	uc             authUseCase
	cookieSecure   bool
	cookieSameSite http.SameSite
}

func NewAuthHandler(uc authUseCase, cookieSecure bool, cookieSameSite http.SameSite) *AuthHandler {
	return &AuthHandler{uc: uc, cookieSecure: cookieSecure, cookieSameSite: cookieSameSite}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"displayName"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ginWriteError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.uc.Register(c.Request.Context(), req.Email, req.Password, req.DisplayName)
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

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email      string `json:"email"`
		Password   string `json:"password"`
		RememberMe bool   `json:"rememberMe"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ginWriteError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	access, refresh, err := h.uc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		ginWriteErr(c, err)
		return
	}

	maxAge := refreshCookieMaxAgeDefault
	if req.RememberMe {
		maxAge = refreshCookieMaxAgeRemember
	}
	c.SetSameSite(h.cookieSameSite)
	c.SetCookie(RefreshCookieName, refresh, maxAge, "/", "", h.cookieSecure, true)

	c.JSON(http.StatusOK, gin.H{
		"accessToken": access,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	cookieVal, err := c.Cookie(RefreshCookieName)
	if err != nil {
		ginWriteError(c, http.StatusUnauthorized, "missing refresh token")
		return
	}

	access, refresh, err := h.uc.RefreshToken(c.Request.Context(), cookieVal)
	if err != nil {
		ginWriteErr(c, err)
		return
	}

	c.SetSameSite(h.cookieSameSite)
	c.SetCookie(RefreshCookieName, refresh, refreshCookieMaxAgeRemember, "/", "", h.cookieSecure, true)

	c.JSON(http.StatusOK, gin.H{
		"accessToken": access,
	})
}

func (h *AuthHandler) SetupStatus(c *gin.Context) {
	adminExists, err := h.uc.SetupStatus(c.Request.Context())
	if err != nil {
		ginWriteErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"adminExists": adminExists,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetSameSite(h.cookieSameSite)
	c.SetCookie(RefreshCookieName, "", -1, "/", "", h.cookieSecure, true)
	c.Status(http.StatusOK)
}
