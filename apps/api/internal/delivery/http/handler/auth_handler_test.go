package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/handler"
	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

// ---- mock usecase ----------------------------------------------------------

type mockAuthUC struct {
	registerFn      func(ctx context.Context, email, password string) (*entity.User, error)
	loginFn         func(ctx context.Context, email, password string) (string, string, error)
	refreshTokenFn  func(ctx context.Context, refreshToken string) (string, error)
	logoutFn        func(ctx context.Context, userID string) error
	setupStatusFn   func(ctx context.Context) (bool, error)
}

func (m *mockAuthUC) Register(ctx context.Context, email, password string) (*entity.User, error) {
	return m.registerFn(ctx, email, password)
}
func (m *mockAuthUC) Login(ctx context.Context, email, password string) (string, string, error) {
	return m.loginFn(ctx, email, password)
}
func (m *mockAuthUC) RefreshToken(ctx context.Context, rt string) (string, error) {
	return m.refreshTokenFn(ctx, rt)
}
func (m *mockAuthUC) Logout(ctx context.Context, userID string) error {
	return m.logoutFn(ctx, userID)
}
func (m *mockAuthUC) SetupStatus(ctx context.Context) (bool, error) {
	if m.setupStatusFn != nil {
		return m.setupStatusFn(ctx)
	}
	return false, nil
}

// ---- helpers ---------------------------------------------------------------

func jsonBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return bytes.NewBuffer(b)
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return out
}

func cookieByName(resp *http.Response, name string) *http.Cookie {
	for _, c := range resp.Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// ---- Register --------------------------------------------------------------

func TestRegisterHandler(t *testing.T) {
	tests := []struct {
		name       string
		uc         *mockAuthUC
		wantStatus int
	}{
		{
			name: "success → 201 with user JSON",
			uc: &mockAuthUC{
				registerFn: func(_ context.Context, email, _ string) (*entity.User, error) {
					return &entity.User{ID: "u1", Email: email, Role: entity.RoleGuest}, nil
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "conflict → 409",
			uc: &mockAuthUC{
				registerFn: func(_ context.Context, _, _ string) (*entity.User, error) {
					return nil, pkgerrors.ErrConflict
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:       "bad JSON body → 400",
			uc:         &mockAuthUC{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewAuthHandler(tc.uc)

			var body *bytes.Buffer
			if tc.name == "bad JSON body → 400" {
				body = bytes.NewBufferString("not-json")
			} else {
				body = jsonBody(t, map[string]string{"email": "a@b.com", "password": "secret"})
			}

			r := httptest.NewRequest(http.MethodPost, "/auth/register", body)
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.Register(w, r)

			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", w.Code, tc.wantStatus, w.Body.String())
			}
			if tc.wantStatus == http.StatusCreated {
				out := decodeBody(t, w)
				if out["id"] == nil || out["email"] == nil || out["role"] == nil {
					t.Errorf("response missing fields: %v", out)
				}
			}
		})
	}
}

// ---- Login -----------------------------------------------------------------

func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name       string
		uc         *mockAuthUC
		wantStatus int
		wantCookie bool
	}{
		{
			name: "success → 200, access token in body, refresh cookie set",
			uc: &mockAuthUC{
				loginFn: func(_ context.Context, _, _ string) (string, string, error) {
					return "access-tok", "refresh-tok", nil
				},
			},
			wantStatus: http.StatusOK,
			wantCookie: true,
		},
		{
			name: "bad credentials → 401",
			uc: &mockAuthUC{
				loginFn: func(_ context.Context, _, _ string) (string, string, error) {
					return "", "", pkgerrors.ErrUnauthorized
				},
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "bad JSON → 400",
			uc:         &mockAuthUC{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewAuthHandler(tc.uc)

			var body *bytes.Buffer
			if tc.name == "bad JSON → 400" {
				body = bytes.NewBufferString("{bad}")
			} else {
				body = jsonBody(t, map[string]string{"email": "a@b.com", "password": "pass"})
			}

			r := httptest.NewRequest(http.MethodPost, "/auth/login", body)
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.Login(w, r)

			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", w.Code, tc.wantStatus, w.Body.String())
			}
			if tc.wantStatus == http.StatusOK {
				out := decodeBody(t, w)
				if out["accessToken"] == nil {
					t.Errorf("response missing accessToken: %v", out)
				}
			}
			if tc.wantCookie {
				resp := w.Result()
				c := cookieByName(resp, handler.RefreshCookieName)
				if c == nil {
					t.Error("expected refresh_token cookie, got none")
				} else if !c.HttpOnly {
					t.Error("refresh_token cookie must be HttpOnly")
				} else if c.Value != "refresh-tok" {
					t.Errorf("cookie value = %q, want %q", c.Value, "refresh-tok")
				}
			}
		})
	}
}

// ---- Refresh ---------------------------------------------------------------

func TestRefreshHandler(t *testing.T) {
	tests := []struct {
		name       string
		cookie     *http.Cookie
		uc         *mockAuthUC
		wantStatus int
	}{
		{
			name:   "success → 200 with new access token",
			cookie: &http.Cookie{Name: handler.RefreshCookieName, Value: "valid-refresh"},
			uc: &mockAuthUC{
				refreshTokenFn: func(_ context.Context, _ string) (string, error) {
					return "new-access-tok", nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing cookie → 401",
			uc:         &mockAuthUC{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "invalid refresh token → 401",
			cookie: &http.Cookie{Name: handler.RefreshCookieName, Value: "bad"},
			uc: &mockAuthUC{
				refreshTokenFn: func(_ context.Context, _ string) (string, error) {
					return "", pkgerrors.ErrUnauthorized
				},
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewAuthHandler(tc.uc)

			r := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
			if tc.cookie != nil {
				r.AddCookie(tc.cookie)
			}
			w := httptest.NewRecorder()

			h.Refresh(w, r)

			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", w.Code, tc.wantStatus, w.Body.String())
			}
			if tc.wantStatus == http.StatusOK {
				out := decodeBody(t, w)
				if out["accessToken"] == nil {
					t.Errorf("response missing accessToken: %v", out)
				}
			}
		})
	}
}

// ---- SetupStatus -----------------------------------------------------------

func TestSetupStatusHandler(t *testing.T) {
	tests := []struct {
		name       string
		uc         *mockAuthUC
		wantStatus int
		wantExists bool
	}{
		{
			name: "no admin — adminExists false",
			uc: &mockAuthUC{
				setupStatusFn: func(_ context.Context) (bool, error) { return false, nil },
			},
			wantStatus: http.StatusOK,
			wantExists: false,
		},
		{
			name: "admin exists — adminExists true",
			uc: &mockAuthUC{
				setupStatusFn: func(_ context.Context) (bool, error) { return true, nil },
			},
			wantStatus: http.StatusOK,
			wantExists: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewAuthHandler(tc.uc)

			r := httptest.NewRequest(http.MethodGet, "/auth/setup", nil)
			w := httptest.NewRecorder()

			h.SetupStatus(w, r)

			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", w.Code, tc.wantStatus, w.Body.String())
			}
			out := decodeBody(t, w)
			got, _ := out["adminExists"].(bool)
			if got != tc.wantExists {
				t.Errorf("adminExists = %v, want %v", got, tc.wantExists)
			}
		})
	}
}

// ---- Logout ----------------------------------------------------------------

func TestLogoutHandler(t *testing.T) {
	h := handler.NewAuthHandler(&mockAuthUC{
		logoutFn: func(_ context.Context, _ string) error { return nil },
	})

	r := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	w := httptest.NewRecorder()

	h.Logout(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	c := cookieByName(w.Result(), handler.RefreshCookieName)
	if c == nil {
		t.Fatal("expected refresh_token cookie to be cleared")
	}
	if c.MaxAge != -1 {
		t.Errorf("cookie MaxAge = %d, want -1 (clear)", c.MaxAge)
	}
}
