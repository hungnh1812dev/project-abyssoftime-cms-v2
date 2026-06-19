package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

const RefreshCookieName = "refresh_token"

const refreshCookieMaxAge = 7 * 24 * 60 * 60 // 7 days

type authUseCase interface {
	Register(ctx context.Context, email, password string) (*entity.User, error)
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	Logout(ctx context.Context, userID string) error
	SetupStatus(ctx context.Context) (bool, error)
}

type AuthHandler struct {
	uc authUseCase
}

func NewAuthHandler(uc authUseCase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ginWriteError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.uc.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		ginWriteErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	c.SetCookie(RefreshCookieName, refresh, refreshCookieMaxAge, "/", "", false, true)

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

	access, err := h.uc.RefreshToken(c.Request.Context(), cookieVal)
	if err != nil {
		ginWriteErr(c, err)
		return
	}

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
	c.SetCookie(RefreshCookieName, "", -1, "/", "", false, true)
	c.Status(http.StatusOK)
}

// ---- response helpers -------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeErr(w http.ResponseWriter, err error) {
	switch {
	case pkgerrors.Is(err, pkgerrors.ErrConflict):
		writeError(w, http.StatusConflict, err.Error())
	case pkgerrors.Is(err, pkgerrors.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, err.Error())
	case pkgerrors.Is(err, pkgerrors.ErrForbidden):
		writeError(w, http.StatusForbidden, err.Error())
	case pkgerrors.Is(err, pkgerrors.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case pkgerrors.Is(err, pkgerrors.ErrBadRequest):
		writeError(w, http.StatusBadRequest, err.Error())
	case pkgerrors.Is(err, pkgerrors.ErrValidation):
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
