package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/handler"
)

func TestLocaleHandler_List_ReturnsSupportedLocales(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.NewLocaleHandler([]string{"en", "vi"})

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/api/locales", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/locales", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("List() status = %d, want 200", w.Code)
	}
	var out []string
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	want := []string{"en", "vi"}
	if len(out) != len(want) {
		t.Fatalf("List() = %v, want %v", out, want)
	}
	for i := range want {
		if out[i] != want[i] {
			t.Errorf("List()[%d] = %q, want %q", i, out[i], want[i])
		}
	}
}
