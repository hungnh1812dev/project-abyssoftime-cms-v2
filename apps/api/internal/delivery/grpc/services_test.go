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

// --- auth mock ---

type mockAuthUC struct {
	registerFn    func(ctx context.Context, email, password, displayName string) (*entity.User, error)
	loginFn       func(ctx context.Context, email, password string) (string, string, error)
	refreshFn     func(ctx context.Context, rt string) (string, string, error)
	setupStatusFn func(ctx context.Context) (bool, error)
}

func (m *mockAuthUC) Register(ctx context.Context, email, password, displayName string) (*entity.User, error) {
	return m.registerFn(ctx, email, password, displayName)
}
func (m *mockAuthUC) Login(ctx context.Context, email, password string) (string, string, error) {
	return m.loginFn(ctx, email, password)
}
func (m *mockAuthUC) RefreshToken(ctx context.Context, rt string) (string, string, error) {
	return m.refreshFn(ctx, rt)
}
func (m *mockAuthUC) SetupStatus(ctx context.Context) (bool, error) {
	return m.setupStatusFn(ctx)
}

// --- content-type mock ---

type mockCTUC struct {
	findBySlugFn func(ctx context.Context, slug string) (*entity.ContentType, error)
	findAllFn    func(ctx context.Context) ([]*entity.ContentType, error)
}

func (m *mockCTUC) FindBySlug(ctx context.Context, slug string) (*entity.ContentType, error) {
	return m.findBySlugFn(ctx, slug)
}
func (m *mockCTUC) FindAll(ctx context.Context) ([]*entity.ContentType, error) {
	return m.findAllFn(ctx)
}

// --- auth service tests ---

func TestAuthService_Login_OK(t *testing.T) {
	uc := &mockAuthUC{
		loginFn: func(_ context.Context, _, _ string) (string, string, error) {
			return "access", "refresh", nil
		},
	}
	svc := NewAuthServiceServer(uc)
	resp, err := svc.Login(context.Background(), &pb.LoginRequest{Email: "a@b.com", Password: "pass"})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if resp.AccessToken != "access" {
		t.Errorf("AccessToken = %q, want %q", resp.AccessToken, "access")
	}
}

func TestAuthService_Login_Unauthorized(t *testing.T) {
	uc := &mockAuthUC{
		loginFn: func(_ context.Context, _, _ string) (string, string, error) {
			return "", "", pkgerrors.ErrUnauthorized
		},
	}
	svc := NewAuthServiceServer(uc)
	_, err := svc.Login(context.Background(), &pb.LoginRequest{})
	if s, ok := status.FromError(err); !ok || s.Code() != codes.Unauthenticated {
		t.Errorf("code = %v, want Unauthenticated", err)
	}
}

func TestAuthService_Register_OK(t *testing.T) {
	uc := &mockAuthUC{
		registerFn: func(_ context.Context, email, _, _ string) (*entity.User, error) {
			return &entity.User{DocumentID: "u1", Email: email, Role: entity.RoleAdmin}, nil
		},
	}
	svc := NewAuthServiceServer(uc)
	resp, err := svc.Register(context.Background(), &pb.RegisterRequest{Email: "a@b.com", Password: "pass"})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if resp.Id != "u1" {
		t.Errorf("Id = %q, want %q", resp.Id, "u1")
	}
}

func TestAuthService_Register_Conflict(t *testing.T) {
	uc := &mockAuthUC{
		registerFn: func(_ context.Context, _, _, _ string) (*entity.User, error) {
			return nil, pkgerrors.ErrConflict
		},
	}
	svc := NewAuthServiceServer(uc)
	_, err := svc.Register(context.Background(), &pb.RegisterRequest{})
	if s, ok := status.FromError(err); !ok || s.Code() != codes.AlreadyExists {
		t.Errorf("code = %v, want AlreadyExists", err)
	}
}

func TestAuthService_SetupStatus(t *testing.T) {
	uc := &mockAuthUC{
		setupStatusFn: func(_ context.Context) (bool, error) { return true, nil },
	}
	svc := NewAuthServiceServer(uc)
	resp, err := svc.SetupStatus(context.Background(), &pb.SetupStatusRequest{})
	if err != nil {
		t.Fatalf("SetupStatus: %v", err)
	}
	if !resp.HasAdmin {
		t.Error("HasAdmin = false, want true")
	}
}

// --- content-type service tests ---

func TestContentTypeService_ListContentTypes(t *testing.T) {
	uc := &mockCTUC{
		findAllFn: func(_ context.Context) ([]*entity.ContentType, error) {
			return []*entity.ContentType{
				{DocumentID: "1", Name: "Blog", Slug: "blog", Kind: entity.KindCollection},
			}, nil
		},
	}
	svc := NewContentTypeServiceServer(uc)
	resp, err := svc.ListContentTypes(context.Background(), &pb.ListContentTypesRequest{})
	if err != nil {
		t.Fatalf("ListContentTypes: %v", err)
	}
	if len(resp.ContentTypes) != 1 {
		t.Fatalf("count = %d, want 1", len(resp.ContentTypes))
	}
	if resp.ContentTypes[0].Slug != "blog" {
		t.Errorf("slug = %q, want %q", resp.ContentTypes[0].Slug, "blog")
	}
}

func TestContentTypeService_GetContentType_OK(t *testing.T) {
	uc := &mockCTUC{
		findBySlugFn: func(_ context.Context, slug string) (*entity.ContentType, error) {
			return &entity.ContentType{DocumentID: "1", Slug: slug, Kind: entity.KindSingle}, nil
		},
	}
	svc := NewContentTypeServiceServer(uc)
	resp, err := svc.GetContentType(context.Background(), &pb.GetContentTypeRequest{Slug: "about"})
	if err != nil {
		t.Fatalf("GetContentType: %v", err)
	}
	if resp.Slug != "about" {
		t.Errorf("slug = %q, want %q", resp.Slug, "about")
	}
}

func TestContentTypeService_GetContentType_NotFound(t *testing.T) {
	uc := &mockCTUC{
		findBySlugFn: func(_ context.Context, _ string) (*entity.ContentType, error) {
			return nil, pkgerrors.ErrNotFound
		},
	}
	svc := NewContentTypeServiceServer(uc)
	_, err := svc.GetContentType(context.Background(), &pb.GetContentTypeRequest{Slug: "nope"})
	if s, ok := status.FromError(err); !ok || s.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", err)
	}
}
