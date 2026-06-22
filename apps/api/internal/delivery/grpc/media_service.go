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
	usecase mediaUseCase
}

func NewMediaServiceServer(usecase mediaUseCase) *MediaServiceServer {
	return &MediaServiceServer{usecase: usecase}
}

func (server *MediaServiceServer) ListMedia(ctx context.Context, req *pb.ListMediaRequest) (*pb.ListMediaResponse, error) {
	items, total, err := server.usecase.List(ctx, int(req.Page), int(req.Limit))
	if err != nil {
		return nil, toGRPCError(err)
	}
	assets := make([]*pb.MediaAsset, len(items))
	for i, asset := range items {
		assets[i] = toProtoMedia(asset)
	}
	return &pb.ListMediaResponse{Assets: assets, Total: total}, nil
}

func (server *MediaServiceServer) DeleteMedia(ctx context.Context, req *pb.DeleteMediaRequest) (*pb.DeleteMediaResponse, error) {
	if err := server.usecase.Delete(ctx, req.Id); err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.DeleteMediaResponse{Success: true}, nil
}

func toProtoMedia(mediaAsset *entity.MediaAsset) *pb.MediaAsset {
	return &pb.MediaAsset{
		Id:           mediaAsset.DocumentID,
		DocumentId:   mediaAsset.DocumentID,
		Url:          mediaAsset.URL,
		ThumbnailUrl: mediaAsset.ThumbnailURL,
		PublicId:     mediaAsset.PublicID,
		FileName:     mediaAsset.FileName,
		FileExt:      mediaAsset.FileExt,
		Hash:         mediaAsset.Hash,
		CreatedAt:    timestamppb.New(mediaAsset.CreatedAt),
	}
}
