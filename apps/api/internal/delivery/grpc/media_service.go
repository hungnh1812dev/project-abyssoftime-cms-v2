package grpcdelivery

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pb "project-abyssoftime-cms-v2/api/proto/cms/v1"
)

type mediaUseCase interface {
	List(ctx context.Context, page, limit int) ([]*entity.MediaAsset, int64, error)
	Delete(ctx context.Context, id string) error
}

type MediaServiceServer struct {
	pb.UnimplementedMediaServiceServer
	uc mediaUseCase
}

func NewMediaServiceServer(uc mediaUseCase) *MediaServiceServer {
	return &MediaServiceServer{uc: uc}
}

func (s *MediaServiceServer) ListMedia(ctx context.Context, req *pb.ListMediaRequest) (*pb.ListMediaResponse, error) {
	items, total, err := s.uc.List(ctx, int(req.Page), int(req.Limit))
	if err != nil {
		return nil, toGRPCError(err)
	}
	assets := make([]*pb.MediaAsset, len(items))
	for i, a := range items {
		assets[i] = toProtoMedia(a)
	}
	return &pb.ListMediaResponse{Assets: assets, Total: total}, nil
}

func (s *MediaServiceServer) DeleteMedia(ctx context.Context, req *pb.DeleteMediaRequest) (*pb.DeleteMediaResponse, error) {
	if err := s.uc.Delete(ctx, req.Id); err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.DeleteMediaResponse{Success: true}, nil
}

func toProtoMedia(a *entity.MediaAsset) *pb.MediaAsset {
	return &pb.MediaAsset{
		Id:           a.DocumentID,
		DocumentId:   a.DocumentID,
		Url:          a.URL,
		ThumbnailUrl: a.ThumbnailURL,
		PublicId:     a.PublicID,
		FileName:     a.FileName,
		FileExt:      a.FileExt,
		Hash:         a.Hash,
		CreatedAt:    timestamppb.New(a.CreatedAt),
	}
}
