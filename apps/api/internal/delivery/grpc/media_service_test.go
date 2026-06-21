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

type mockMediaUC struct {
	listFn   func(ctx context.Context, page, limit int) ([]*entity.MediaAsset, int64, error)
	deleteFn func(ctx context.Context, id string) error
}

func (m *mockMediaUC) List(ctx context.Context, page, limit int) ([]*entity.MediaAsset, int64, error) {
	return m.listFn(ctx, page, limit)
}
func (m *mockMediaUC) Delete(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}

func TestMediaService_ListMedia_OK(t *testing.T) {
	uc := &mockMediaUC{
		listFn: func(_ context.Context, _, _ int) ([]*entity.MediaAsset, int64, error) {
			return []*entity.MediaAsset{{DocumentID: "m1", URL: "https://cdn/a.jpg"}}, 3, nil
		},
	}
	svc := NewMediaServiceServer(uc)
	resp, err := svc.ListMedia(context.Background(), &pb.ListMediaRequest{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("ListMedia: %v", err)
	}
	if resp.Total != 3 {
		t.Errorf("Total = %d, want 3", resp.Total)
	}
	if len(resp.Assets) != 1 {
		t.Errorf("Assets count = %d, want 1", len(resp.Assets))
	}
	if resp.Assets[0].Url != "https://cdn/a.jpg" {
		t.Errorf("URL = %q", resp.Assets[0].Url)
	}
}

func TestMediaService_DeleteMedia_OK(t *testing.T) {
	uc := &mockMediaUC{
		deleteFn: func(_ context.Context, _ string) error { return nil },
	}
	svc := NewMediaServiceServer(uc)
	resp, err := svc.DeleteMedia(context.Background(), &pb.DeleteMediaRequest{Id: "m1"})
	if err != nil {
		t.Fatalf("DeleteMedia: %v", err)
	}
	if !resp.Success {
		t.Error("Success = false, want true")
	}
}

func TestMediaService_DeleteMedia_NotFound(t *testing.T) {
	uc := &mockMediaUC{
		deleteFn: func(_ context.Context, _ string) error { return pkgerrors.ErrNotFound },
	}
	svc := NewMediaServiceServer(uc)
	_, err := svc.DeleteMedia(context.Background(), &pb.DeleteMediaRequest{Id: "nope"})
	if s, ok := status.FromError(err); !ok || s.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", err)
	}
}
