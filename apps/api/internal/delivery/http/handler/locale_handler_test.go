package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/handler"
)

func TestLocaleHandler_List_ReturnsSupportedLocales(t *testing.T) {
	h := handler.NewLocaleHandler([]string{"en", "vi"})

	req := httptest.NewRequest(http.MethodGet, "/api/locales", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

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
