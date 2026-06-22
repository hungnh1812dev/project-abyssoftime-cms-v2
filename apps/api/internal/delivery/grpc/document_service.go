package grpcdelivery

import (
	"context"
	"encoding/json"

	"google.golang.org/protobuf/types/known/timestamppb"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/middleware"
	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pb "project-abyssoftime-cms-v2/api/proto/cms/v1"
)

type documentUseCase interface {
	Save(ctx context.Context, contentTypeSlug string, doc *entity.Document, fields []entity.FieldDefinition, userID string) (*entity.Document, error)
	GetForEdit(ctx context.Context, contentTypeSlug, documentID, locale string, fields []entity.FieldDefinition) (*entity.Document, string, error)
	GetPublished(ctx context.Context, contentTypeSlug, documentID, locale string, fields []entity.FieldDefinition) (*entity.Document, error)
	Publish(ctx context.Context, contentTypeSlug, documentID, locale string, fields []entity.FieldDefinition, userID string) error
	Unpublish(ctx context.Context, contentTypeSlug, documentID, locale string) error
	Delete(ctx context.Context, contentTypeSlug, documentID string, fields []entity.FieldDefinition) error
	GetSingleType(ctx context.Context, contentTypeSlug, locale string, fields []entity.FieldDefinition) (*entity.Document, string, error)
	SaveSingleType(ctx context.Context, contentTypeSlug string, data map[string]any, locale string, fields []entity.FieldDefinition, userID string) (*entity.Document, error)
	PublishSingleType(ctx context.Context, contentTypeSlug, locale string, fields []entity.FieldDefinition, userID string) error
	UnpublishSingleType(ctx context.Context, contentTypeSlug, locale string) error
	GetAllPaginated(ctx context.Context, contentTypeSlug string, start, size int, locale string, fields []entity.FieldDefinition, orderBy string, sortDir int) ([]*entity.Document, []string, int64, error)
}

type DocumentServiceServer struct {
	pb.UnimplementedDocumentServiceServer
	usecase documentUseCase
}

func NewDocumentServiceServer(usecase documentUseCase) *DocumentServiceServer {
	return &DocumentServiceServer{usecase: usecase}
}

func (server *DocumentServiceServer) GetDocument(ctx context.Context, req *pb.GetDocumentRequest) (*pb.Document, error) {
	doc, _, err := server.usecase.GetForEdit(ctx, req.ContentTypeSlug, req.DocumentId, req.Locale, nil)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoDoc(doc), nil
}

func (server *DocumentServiceServer) ListDocuments(ctx context.Context, req *pb.ListDocumentsRequest) (*pb.ListDocumentsResponse, error) {
	docs, _, total, err := server.usecase.GetAllPaginated(ctx, req.ContentTypeSlug, int(req.Start), int(req.Size), req.Locale, nil, "createdAt", -1)
	if err != nil {
		return nil, toGRPCError(err)
	}
	items := make([]*pb.Document, len(docs))
	for i, doc := range docs {
		items[i] = toProtoDoc(doc)
	}
	return &pb.ListDocumentsResponse{Items: items, Total: total, Start: req.Start, Size: req.Size}, nil
}

func (server *DocumentServiceServer) SaveDocument(ctx context.Context, req *pb.SaveDocumentRequest) (*pb.Document, error) {
	data := decodeData(req.Data)
	doc := &entity.Document{DocumentID: req.DocumentId, Fields: data, Locale: req.Locale}
	saved, err := server.usecase.Save(ctx, req.ContentTypeSlug, doc, nil, middleware.UserID(ctx))
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoDoc(saved), nil
}

func (server *DocumentServiceServer) PublishDocument(ctx context.Context, req *pb.PublishDocumentRequest) (*pb.Document, error) {
	if err := server.usecase.Publish(ctx, req.ContentTypeSlug, req.DocumentId, req.Locale, nil, middleware.UserID(ctx)); err != nil {
		return nil, toGRPCError(err)
	}
	doc, err := server.usecase.GetPublished(ctx, req.ContentTypeSlug, req.DocumentId, req.Locale, nil)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoDoc(doc), nil
}

func (server *DocumentServiceServer) UnpublishDocument(ctx context.Context, req *pb.PublishDocumentRequest) (*pb.Document, error) {
	if err := server.usecase.Unpublish(ctx, req.ContentTypeSlug, req.DocumentId, req.Locale); err != nil {
		return nil, toGRPCError(err)
	}
	doc, _, err := server.usecase.GetForEdit(ctx, req.ContentTypeSlug, req.DocumentId, req.Locale, nil)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoDoc(doc), nil
}

func (server *DocumentServiceServer) DeleteDocument(ctx context.Context, req *pb.DeleteDocumentRequest) (*pb.DeleteDocumentResponse, error) {
	if err := server.usecase.Delete(ctx, req.ContentTypeSlug, req.DocumentId, nil); err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.DeleteDocumentResponse{Success: true}, nil
}

func (server *DocumentServiceServer) GetSingleType(ctx context.Context, req *pb.GetSingleTypeRequest) (*pb.Document, error) {
	doc, _, err := server.usecase.GetSingleType(ctx, req.ContentTypeSlug, req.Locale, nil)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoDoc(doc), nil
}

func (server *DocumentServiceServer) SaveSingleType(ctx context.Context, req *pb.SaveSingleTypeRequest) (*pb.Document, error) {
	data := decodeData(req.Data)
	saved, err := server.usecase.SaveSingleType(ctx, req.ContentTypeSlug, data, req.Locale, nil, middleware.UserID(ctx))
	if err != nil {
		return nil, toGRPCError(err)
	}
	doc, _, err := server.usecase.GetSingleType(ctx, req.ContentTypeSlug, saved.Locale, nil)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoDoc(doc), nil
}

func (server *DocumentServiceServer) PublishSingleType(ctx context.Context, req *pb.GetSingleTypeRequest) (*pb.Document, error) {
	if err := server.usecase.PublishSingleType(ctx, req.ContentTypeSlug, req.Locale, nil, middleware.UserID(ctx)); err != nil {
		return nil, toGRPCError(err)
	}
	doc, _, err := server.usecase.GetSingleType(ctx, req.ContentTypeSlug, req.Locale, nil)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoDoc(doc), nil
}

func (server *DocumentServiceServer) UnpublishSingleType(ctx context.Context, req *pb.GetSingleTypeRequest) (*pb.Document, error) {
	if err := server.usecase.UnpublishSingleType(ctx, req.ContentTypeSlug, req.Locale); err != nil {
		return nil, toGRPCError(err)
	}
	doc, _, err := server.usecase.GetSingleType(ctx, req.ContentTypeSlug, req.Locale, nil)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoDoc(doc), nil
}

func toProtoDoc(document *entity.Document) *pb.Document {
	protoBuffer := &pb.Document{
		DocumentId:  document.DocumentID,
		Version:     string(document.Version),
		Data:        encodeData(document.Fields),
		Locale:      document.Locale,
		CreatedAt:   timestamppb.New(document.CreatedAt),
		UpdatedAt:   timestamppb.New(document.UpdatedAt),
		CreatedBy:   document.CreatedBy,
		UpdatedBy:   document.UpdatedBy,
		PublishedBy: document.PublishedBy,
	}
	if document.PublishedAt != nil {
		protoBuffer.PublishedAt = timestamppb.New(*document.PublishedAt)
	}
	return protoBuffer
}

func encodeData(data map[string]any) []byte {
	if data == nil {
		return nil
	}
	b, _ := json.Marshal(data)
	return b
}

func decodeData(b []byte) map[string]any {
	if len(b) == 0 {
		return nil
	}
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	return m
}
