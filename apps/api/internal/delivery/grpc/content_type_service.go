package grpcdelivery

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pb "project-abyssoftime-cms-v2/api/proto/cms/v1"
)

type contentTypeUseCase interface {
	FindBySlug(ctx context.Context, slug string) (*entity.ContentType, error)
	FindAll(ctx context.Context) ([]*entity.ContentType, error)
}

type ContentTypeServiceServer struct {
	pb.UnimplementedContentTypeServiceServer
	uc contentTypeUseCase
}

func NewContentTypeServiceServer(uc contentTypeUseCase) *ContentTypeServiceServer {
	return &ContentTypeServiceServer{uc: uc}
}

func (s *ContentTypeServiceServer) GetContentType(ctx context.Context, req *pb.GetContentTypeRequest) (*pb.ContentType, error) {
	ct, err := s.uc.FindBySlug(ctx, req.Slug)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoCT(ct), nil
}

func (s *ContentTypeServiceServer) ListContentTypes(ctx context.Context, _ *pb.ListContentTypesRequest) (*pb.ListContentTypesResponse, error) {
	cts, err := s.uc.FindAll(ctx)
	if err != nil {
		return nil, toGRPCError(err)
	}
	result := make([]*pb.ContentType, len(cts))
	for i, ct := range cts {
		result[i] = toProtoCT(ct)
	}
	return &pb.ListContentTypesResponse{ContentTypes: result}, nil
}

func toProtoCT(ct *entity.ContentType) *pb.ContentType {
	return &pb.ContentType{
		Id:         ct.DocumentID,
		Name:       ct.Name,
		Slug:       ct.Slug,
		Kind:       string(ct.Kind),
		Fields:     toProtoFields(ct.Fields),
		ListFields: ct.ListFields,
		CreatedAt:  timestamppb.New(ct.CreatedAt),
		UpdatedAt:  timestamppb.New(ct.UpdatedAt),
	}
}

func toProtoFields(fields []entity.FieldDefinition) []*pb.FieldDefinition {
	if len(fields) == 0 {
		return nil
	}
	result := make([]*pb.FieldDefinition, len(fields))
	for i, f := range fields {
		result[i] = &pb.FieldDefinition{
			Name:   f.Name,
			Type:   f.Type,
			Ext:    f.Ext,
			Fields: toProtoFields(f.Fields),
		}
	}
	return result
}
