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
	usecase contentTypeUseCase
}

func NewContentTypeServiceServer(usecase contentTypeUseCase) *ContentTypeServiceServer {
	return &ContentTypeServiceServer{usecase: usecase}
}

func (server *ContentTypeServiceServer) GetContentType(ctx context.Context, req *pb.GetContentTypeRequest) (*pb.ContentType, error) {
	contentType, err := server.usecase.FindBySlug(ctx, req.Slug)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoCT(contentType), nil
}

func (server *ContentTypeServiceServer) ListContentTypes(ctx context.Context, _ *pb.ListContentTypesRequest) (*pb.ListContentTypesResponse, error) {
	contentTypes, err := server.usecase.FindAll(ctx)
	if err != nil {
		return nil, toGRPCError(err)
	}
	result := make([]*pb.ContentType, len(contentTypes))
	for i, contentType := range contentTypes {
		result[i] = toProtoCT(contentType)
	}
	return &pb.ListContentTypesResponse{ContentTypes: result}, nil
}

func toProtoCT(contentType *entity.ContentType) *pb.ContentType {
	return &pb.ContentType{
		Id:         contentType.DocumentID,
		Name:       contentType.Name,
		Slug:       contentType.Slug,
		Kind:       string(contentType.Kind),
		Fields:     toProtoFields(contentType.Fields),
		ListFields: contentType.ListFields,
		CreatedAt:  timestamppb.New(contentType.CreatedAt),
		UpdatedAt:  timestamppb.New(contentType.UpdatedAt),
	}
}

func toProtoFields(fields []entity.FieldDefinition) []*pb.FieldDefinition {
	if len(fields) == 0 {
		return nil
	}
	result := make([]*pb.FieldDefinition, len(fields))
	for i, field := range fields {
		result[i] = &pb.FieldDefinition{
			Name:   field.Name,
			Type:   field.Type,
			Ext:    field.Ext,
			Fields: toProtoFields(field.Fields),
		}
	}
	return result
}
