package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

// RefreshCookieName is exported so tests can reference it without hardcoding.
const RefreshCookieName = "refresh_token"

const refreshCookieMaxAge = 7 * 24 * 60 * 60 // 7 days

type authUseCase interface {
	Register(ctx context.Context, email, password string) (*entity.User, error)
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	Logout(ctx context.Context, userID string) error
}

type AuthHandler struct {
	uc authUseCase
}

func NewAuthHandler(uc authUseCase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.uc.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		writeErr(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	access, refresh, err := h.uc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeErr(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    refresh,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   refreshCookieMaxAge,
		Path:     "/",
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"accessToken": access,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(RefreshCookieName)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	access, err := h.uc.RefreshToken(r.Context(), cookie.Value)
	if err != nil {
		writeErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"accessToken": access,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Path:     "/",
	})
	w.WriteHeader(http.StatusOK)
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
