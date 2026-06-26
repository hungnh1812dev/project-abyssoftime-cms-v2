package grpcdelivery

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
	pb "project-abyssoftime-cms-v2/api/proto/cms/v1"
)

type mockDocUC struct {
	getForEditFn          func(ctx context.Context, slug, docID, locale string) (*entity.Document, string, error)
	getAllPaginatedFn      func(ctx context.Context, slug string, start, size int, locale string, orderBy string, sortDir int) ([]*entity.Document, []string, int64, error)
	saveFn                func(ctx context.Context, slug string, doc *entity.Document, userID string) (*entity.Document, error)
	getPublishedFn        func(ctx context.Context, slug, docID, locale string) (*entity.Document, error)
	publishFn             func(ctx context.Context, slug, docID, locale, userID string) error
	unpublishFn           func(ctx context.Context, slug, docID, locale string) error
	deleteFn              func(ctx context.Context, slug, docID string) error
	getSingleTypeFn       func(ctx context.Context, slug, locale string) (*entity.Document, string, error)
	saveSingleTypeFn      func(ctx context.Context, slug string, data map[string]any, locale, userID string) (*entity.Document, error)
	publishSingleTypeFn   func(ctx context.Context, slug, locale, userID string) error
	unpublishSingleTypeFn func(ctx context.Context, slug, locale string) error
}

func (m *mockDocUC) Save(ctx context.Context, s string, d *entity.Document, _ []entity.FieldDefinition, u string) (*entity.Document, error) {
	return m.saveFn(ctx, s, d, u)
}
func (m *mockDocUC) GetForEdit(ctx context.Context, s, d, l string, _ []entity.FieldDefinition) (*entity.Document, string, error) {
	return m.getForEditFn(ctx, s, d, l)
}
func (m *mockDocUC) GetPublished(ctx context.Context, s, d, l string, _ []entity.FieldDefinition) (*entity.Document, error) {
	return m.getPublishedFn(ctx, s, d, l)
}
func (m *mockDocUC) Publish(ctx context.Context, s, d, l string, _ []entity.FieldDefinition, u string) error {
	return m.publishFn(ctx, s, d, l, u)
}
func (m *mockDocUC) Unpublish(ctx context.Context, s, d, l string, _ []entity.FieldDefinition) error {
	return m.unpublishFn(ctx, s, d, l)
}
func (m *mockDocUC) Delete(ctx context.Context, s, d string, _ []entity.FieldDefinition) error { return m.deleteFn(ctx, s, d) }
func (m *mockDocUC) GetSingleType(ctx context.Context, s, l string, _ []entity.FieldDefinition) (*entity.Document, string, error) {
	return m.getSingleTypeFn(ctx, s, l)
}
func (m *mockDocUC) SaveSingleType(ctx context.Context, s string, data map[string]any, l string, _ []entity.FieldDefinition, u string) (*entity.Document, error) {
	return m.saveSingleTypeFn(ctx, s, data, l, u)
}
func (m *mockDocUC) PublishSingleType(ctx context.Context, s, l string, _ []entity.FieldDefinition, u string) error {
	return m.publishSingleTypeFn(ctx, s, l, u)
}
func (m *mockDocUC) UnpublishSingleType(ctx context.Context, s, l string, _ []entity.FieldDefinition) error {
	return m.unpublishSingleTypeFn(ctx, s, l)
}
func (m *mockDocUC) GetAllPaginated(ctx context.Context, s string, start, size int, l string, _ []entity.FieldDefinition, orderBy string, sortDir int, _ []entity.FilterNode) ([]*entity.Document, []string, int64, error) {
	return m.getAllPaginatedFn(ctx, s, start, size, l, orderBy, sortDir)
}

func TestDocumentService_GetDocument_OK(t *testing.T) {
	uc := &mockDocUC{
		getForEditFn: func(_ context.Context, _, docID, _ string) (*entity.Document, string, error) {
			return &entity.Document{DocumentID: docID, Locale: "en"}, "draft", nil
		},
	}
	svc := NewDocumentServiceServer(uc)
	resp, err := svc.GetDocument(context.Background(), &pb.GetDocumentRequest{
		ContentTypeSlug: "blog", DocumentId: "d1", Locale: "en",
	})
	if err != nil {
		t.Fatalf("GetDocument: %v", err)
	}
	if resp.DocumentId != "d1" {
		t.Errorf("DocumentId = %q, want %q", resp.DocumentId, "d1")
	}
}

func TestDocumentService_GetDocument_NotFound(t *testing.T) {
	uc := &mockDocUC{
		getForEditFn: func(_ context.Context, _, _, _ string) (*entity.Document, string, error) {
			return nil, "", pkgerrors.ErrNotFound
		},
	}
	svc := NewDocumentServiceServer(uc)
	_, err := svc.GetDocument(context.Background(), &pb.GetDocumentRequest{ContentTypeSlug: "blog", DocumentId: "nope"})
	if s, ok := status.FromError(err); !ok || s.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", err)
	}
}

func TestDocumentService_ListDocuments_OK(t *testing.T) {
	uc := &mockDocUC{
		getAllPaginatedFn: func(_ context.Context, _ string, _, _ int, _, _ string, _ int) ([]*entity.Document, []string, int64, error) {
			return []*entity.Document{
				{DocumentID: "d1", Locale: "en"},
			}, []string{"draft"}, 5, nil
		},
	}
	svc := NewDocumentServiceServer(uc)
	resp, err := svc.ListDocuments(context.Background(), &pb.ListDocumentsRequest{
		ContentTypeSlug: "blog", Start: 0, Size: 20, Locale: "en",
	})
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	if resp.Total != 5 {
		t.Errorf("Total = %d, want 5", resp.Total)
	}
	if len(resp.Items) != 1 {
		t.Errorf("Items count = %d, want 1", len(resp.Items))
	}
}

func TestDocumentService_DeleteDocument_OK(t *testing.T) {
	uc := &mockDocUC{
		deleteFn: func(_ context.Context, _, _ string) error { return nil },
	}
	svc := NewDocumentServiceServer(uc)
	resp, err := svc.DeleteDocument(context.Background(), &pb.DeleteDocumentRequest{ContentTypeSlug: "blog", DocumentId: "d1"})
	if err != nil {
		t.Fatalf("DeleteDocument: %v", err)
	}
	if !resp.Success {
		t.Error("Success = false, want true")
	}
}

func TestDocumentService_GetSingleType_OK(t *testing.T) {
	uc := &mockDocUC{
		getSingleTypeFn: func(_ context.Context, _, _ string) (*entity.Document, string, error) {
			return &entity.Document{DocumentID: "s1", Locale: "en"}, "published", nil
		},
	}
	svc := NewDocumentServiceServer(uc)
	resp, err := svc.GetSingleType(context.Background(), &pb.GetSingleTypeRequest{ContentTypeSlug: "about", Locale: "en"})
	if err != nil {
		t.Fatalf("GetSingleType: %v", err)
	}
	if resp.DocumentId != "s1" {
		t.Errorf("DocumentId = %q, want %q", resp.DocumentId, "s1")
	}
}

func TestDocumentService_GetSingleType_NotFound(t *testing.T) {
	uc := &mockDocUC{
		getSingleTypeFn: func(_ context.Context, _, _ string) (*entity.Document, string, error) {
			return nil, "", pkgerrors.ErrNotFound
		},
	}
	svc := NewDocumentServiceServer(uc)
	_, err := svc.GetSingleType(context.Background(), &pb.GetSingleTypeRequest{ContentTypeSlug: "about"})
	if s, ok := status.FromError(err); !ok || s.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", err)
	}
}
