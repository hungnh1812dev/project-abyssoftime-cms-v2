package handler

import (
	"context"
	"net/http"
	"regexp"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

var objectIDRe = regexp.MustCompile(`^[a-f0-9]{24}$`)

type contentTypeUseCase interface {
	FindByID(ctx context.Context, id string) (*entity.ContentType, error)
	FindBySlug(ctx context.Context, slug string) (*entity.ContentType, error)
	FindAll(ctx context.Context) ([]*entity.ContentType, error)
}

type ContentTypeHandler struct {
	uc contentTypeUseCase
}

func NewContentTypeHandler(uc contentTypeUseCase) *ContentTypeHandler {
	return &ContentTypeHandler{uc: uc}
}

type contentTypeSummary struct {
	ID   string             `json:"ID"`
	Name string             `json:"Name"`
	Slug string             `json:"Slug"`
	Kind entity.ContentKind `json:"Kind"`
}

func (h *ContentTypeHandler) ListSummary(w http.ResponseWriter, r *http.Request) {
	cts, err := h.uc.FindAll(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	summaries := make([]contentTypeSummary, len(cts))
	for i, ct := range cts {
		summaries[i] = contentTypeSummary{
			ID:   ct.ID,
			Name: ct.Name,
			Slug: ct.Slug,
			Kind: ct.Kind,
		}
	}
	writeJSON(w, http.StatusOK, summaries)
}

func (h *ContentTypeHandler) Get(w http.ResponseWriter, r *http.Request) {
	identifier := r.PathValue("identifier")
	var (
		ct  *entity.ContentType
		err error
	)
	if objectIDRe.MatchString(identifier) {
		ct, err = h.uc.FindByID(r.Context(), identifier)
	} else {
		ct, err = h.uc.FindBySlug(r.Context(), identifier)
	}
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ct)
}
