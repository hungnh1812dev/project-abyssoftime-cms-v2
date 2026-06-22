package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/handler"
	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	locale "project-abyssoftime-cms-v2/api/internal/usecase/locale"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

type stubLocaleUC struct {
	locales []*entity.Locale
	created *entity.Locale
	updated *entity.Locale
	listErr error
	createErr error
	updateErr error
	deleteErr error
}

func (stub *stubLocaleUC) List(_ context.Context) ([]*entity.Locale, error) {
	return stub.locales, stub.listErr
}

func (stub *stubLocaleUC) Create(_ context.Context, input locale.CreateInput) (*entity.Locale, error) {
	if stub.createErr != nil {
		return nil, stub.createErr
	}
	stub.created = &entity.Locale{Code: input.Code, Name: input.Name, IsDefault: input.IsDefault}
	return stub.created, nil
}

func (stub *stubLocaleUC) Update(_ context.Context, code string, input locale.UpdateInput) (*entity.Locale, error) {
	if stub.updateErr != nil {
		return nil, stub.updateErr
	}
	name := "English"
	if input.Name != nil {
		name = *input.Name
	}
	stub.updated = &entity.Locale{Code: code, Name: name}
	return stub.updated, nil
}

func (stub *stubLocaleUC) Delete(_ context.Context, _ string) error {
	return stub.deleteErr
}

func TestLocaleHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &stubLocaleUC{
		locales: []*entity.Locale{
			{Code: "en", Name: "English", IsDefault: true},
			{Code: "vi", Name: "Tiếng Việt", IsDefault: false},
		},
	}
	handler := handler.NewLocaleHandler(stub)

	recorder := httptest.NewRecorder()
	_, router := gin.CreateTestContext(recorder)
	router.GET("/api/locales", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/api/locales", nil)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}

	var out []entity.Locale
	if err := json.NewDecoder(recorder.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("len = %d, want 2", len(out))
	}
	if out[0].Code != "en" || out[0].Name != "English" || !out[0].IsDefault {
		t.Errorf("first locale = %+v, want en/English/default", out[0])
	}
}

func TestLocaleHandler_Create_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &stubLocaleUC{}
	handler := handler.NewLocaleHandler(stub)

	recorder := httptest.NewRecorder()
	_, router := gin.CreateTestContext(recorder)
	router.POST("/api/locales", handler.Create)

	body := `{"code":"ja","name":"日本語","isDefault":false}`
	req := httptest.NewRequest(http.MethodPost, "/api/locales", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", recorder.Code)
	}
	if stub.created == nil || stub.created.Code != "ja" {
		t.Error("expected locale to be created with code 'ja'")
	}
}

func TestLocaleHandler_Create_Conflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &stubLocaleUC{createErr: pkgerrors.ErrConflict}
	handler := handler.NewLocaleHandler(stub)

	recorder := httptest.NewRecorder()
	_, router := gin.CreateTestContext(recorder)
	router.POST("/api/locales", handler.Create)

	body := `{"code":"en","name":"English"}`
	req := httptest.NewRequest(http.MethodPost, "/api/locales", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", recorder.Code)
	}
}

func TestLocaleHandler_Update_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &stubLocaleUC{}
	handler := handler.NewLocaleHandler(stub)

	recorder := httptest.NewRecorder()
	_, router := gin.CreateTestContext(recorder)
	router.PUT("/api/locales/:code", handler.Update)

	body := `{"name":"English (US)"}`
	req := httptest.NewRequest(http.MethodPut, "/api/locales/en", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
}

func TestLocaleHandler_Delete_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &stubLocaleUC{}
	handler := handler.NewLocaleHandler(stub)

	recorder := httptest.NewRecorder()
	_, router := gin.CreateTestContext(recorder)
	router.DELETE("/api/locales/:code", handler.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api/locales/vi", nil)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", recorder.Code)
	}
}

func TestLocaleHandler_Delete_Conflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &stubLocaleUC{deleteErr: pkgerrors.ErrConflict}
	handler := handler.NewLocaleHandler(stub)

	recorder := httptest.NewRecorder()
	_, router := gin.CreateTestContext(recorder)
	router.DELETE("/api/locales/:code", handler.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api/locales/en", nil)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", recorder.Code)
	}
}
