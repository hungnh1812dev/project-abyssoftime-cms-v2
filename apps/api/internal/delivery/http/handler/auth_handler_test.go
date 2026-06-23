package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/handler"
	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

type mockAuthUC struct {
	registerFn     func(ctx context.Context, email, password, displayName string) (*entity.User, error)
	loginFn        func(ctx context.Context, email, password string) (string, string, error)
	refreshTokenFn func(ctx context.Context, refreshToken string) (string, string, error)
	logoutFn       func(ctx context.Context, userID string) error
	setupStatusFn  func(ctx context.Context) (bool, error)
}

func (m *mockAuthUC) Register(ctx context.Context, email, password, displayName string) (*entity.User, error) {
	return m.registerFn(ctx, email, password, displayName)
}
func (m *mockAuthUC) Login(ctx context.Context, email, password string) (string, string, error) {
	return m.loginFn(ctx, email, password)
}
func (m *mockAuthUC) RefreshToken(ctx context.Context, rt string) (string, string, error) {
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

func TestRegisterHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		uc         *mockAuthUC
		body       string
		wantStatus int
	}{
		{
			name: "success → 201 with user JSON",
			uc: &mockAuthUC{
				registerFn: func(_ context.Context, email, _, displayName string) (*entity.User, error) {
					return &entity.User{DocumentID: "u1", Email: email, DisplayName: displayName, Role: entity.RoleGuest}, nil
				},
			},
			body:       `{"email":"a@b.com","password":"secret","displayName":"Test User"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name: "conflict → 409",
			uc: &mockAuthUC{
				registerFn: func(_ context.Context, _, _, _ string) (*entity.User, error) {
					return nil, pkgerrors.ErrConflict
				},
			},
			body:       `{"email":"a@b.com","password":"secret","displayName":"Test User"}`,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "bad JSON body → 400",
			uc:         &mockAuthUC{},
			body:       "not-json",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewAuthHandler(tc.uc, false, http.SameSiteLaxMode)
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/auth/register", h.Register)

			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

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

func TestLoginHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		uc         *mockAuthUC
		body       string
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
			body:       `{"email":"a@b.com","password":"pass"}`,
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
			body:       `{"email":"a@b.com","password":"pass"}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "bad JSON → 400",
			uc:         &mockAuthUC{},
			body:       "{bad}",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewAuthHandler(tc.uc, false, http.SameSiteLaxMode)
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/auth/login", h.Login)

			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", w.Code, tc.wantStatus, w.Body.String())
			}
			if tc.wantStatus == http.StatusOK {
				out := decodeBody(t, w)
				if out["accessToken"] == nil {
					t.Errorf("response missing accessToken: %v", out)
				}
				if out["refreshToken"] == nil {
					t.Errorf("response missing refreshToken: %v", out)
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

func TestRefreshHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	successUC := &mockAuthUC{
		refreshTokenFn: func(_ context.Context, token string) (string, string, error) {
			return "new-access-tok", "new-refresh-tok", nil
		},
	}

	tests := []struct {
		name       string
		cookie     *http.Cookie
		body       string
		uc         *mockAuthUC
		wantStatus int
	}{
		{
			name:   "success with cookie → 200",
			cookie: &http.Cookie{Name: handler.RefreshCookieName, Value: "valid-refresh"},
			uc:     successUC,

			wantStatus: http.StatusOK,
		},
		{
			name:       "success with body token → 200",
			body:       `{"refreshToken":"valid-refresh"}`,
			uc:         successUC,
			wantStatus: http.StatusOK,
		},
		{
			name:   "body token takes precedence over cookie",
			cookie: &http.Cookie{Name: handler.RefreshCookieName, Value: "cookie-tok"},
			body:   `{"refreshToken":"body-tok"}`,
			uc: &mockAuthUC{
				refreshTokenFn: func(_ context.Context, token string) (string, string, error) {
					if token != "body-tok" {
						return "", "", pkgerrors.ErrUnauthorized
					}
					return "new-access-tok", "new-refresh-tok", nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no cookie and no body → 401",
			uc:         &mockAuthUC{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "invalid refresh token via cookie → 401",
			cookie: &http.Cookie{Name: handler.RefreshCookieName, Value: "bad"},
			uc: &mockAuthUC{
				refreshTokenFn: func(_ context.Context, _ string) (string, string, error) {
					return "", "", pkgerrors.ErrUnauthorized
				},
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid refresh token via body → 401",
			body: `{"refreshToken":"bad"}`,
			uc: &mockAuthUC{
				refreshTokenFn: func(_ context.Context, _ string) (string, string, error) {
					return "", "", pkgerrors.ErrUnauthorized
				},
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewAuthHandler(tc.uc, false, http.SameSiteLaxMode)
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/auth/refresh", h.Refresh)

			var bodyReader *bytes.Buffer
			if tc.body != "" {
				bodyReader = bytes.NewBufferString(tc.body)
			} else {
				bodyReader = &bytes.Buffer{}
			}
			req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bodyReader)
			req.Header.Set("Content-Type", "application/json")
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}
			r.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", w.Code, tc.wantStatus, w.Body.String())
			}
			if tc.wantStatus == http.StatusOK {
				out := decodeBody(t, w)
				if out["accessToken"] == nil {
					t.Errorf("response missing accessToken: %v", out)
				}
				if out["refreshToken"] == nil {
					t.Errorf("response missing refreshToken: %v", out)
				}
				resp := w.Result()
				cookie := cookieByName(resp, handler.RefreshCookieName)
				if cookie == nil {
					t.Error("expected refresh_token cookie to be re-issued")
				} else if cookie.Value != "new-refresh-tok" {
					t.Errorf("cookie value = %q, want %q", cookie.Value, "new-refresh-tok")
				}
			}
		})
	}
}

func TestSetupStatusHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

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
			h := handler.NewAuthHandler(tc.uc, false, http.SameSiteLaxMode)
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.GET("/auth/setup", h.SetupStatus)

			req := httptest.NewRequest(http.MethodGet, "/auth/setup", nil)
			r.ServeHTTP(w, req)

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

func TestLogoutHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := handler.NewAuthHandler(&mockAuthUC{
		logoutFn: func(_ context.Context, _ string) error { return nil },
	}, false, http.SameSiteLaxMode)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/auth/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	r.ServeHTTP(w, req)

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
