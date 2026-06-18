package handler

import "net/http"

// LocaleHandler exposes the server's configured locale list — the FE's
// single source of truth for the locale switcher.
type LocaleHandler struct {
	supportedLocales []string
}

func NewLocaleHandler(supportedLocales []string) *LocaleHandler {
	return &LocaleHandler{supportedLocales: supportedLocales}
}

func (h *LocaleHandler) List(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.supportedLocales)
}
