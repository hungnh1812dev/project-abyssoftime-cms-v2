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
	GetAllPaginated(ctx context.Context, contentTypeSlug string, start, size int, locale string, fields []entity.FieldDefinition) ([]*entity.Document, []string, int64, error)
}

type DocumentServiceServer struct {
	pb.UnimplementedDocumentServiceServer
	uc documentUseCase
}

func NewDocumentServiceServer(uc documentUseCase) *DocumentServiceServer {
	return &DocumentServiceServer{uc: uc}
}

func (s *DocumentServiceServer) GetDocument(ctx context.Context, req *pb.GetDocumentRequest) (*pb.Document, error) {
	doc, _, err := s.uc.GetForEdit(ctx, req.ContentTypeSlug, req.DocumentId, req.Locale, nil)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoDoc(doc), nil
}

func (s *DocumentServiceServer) ListDocuments(ctx context.Context, req *pb.ListDocumentsRequest) (*pb.ListDocumentsResponse, error) {
	docs, _, total, err := s.uc.GetAllPaginated(ctx, req.ContentTypeSlug, int(req.Start), int(req.Size), req.Locale, nil)
	if err != nil {
		return nil, toGRPCError(err)
	}
	items := make([]*pb.Document, len(docs))
	for i, d := range docs {
		items[i] = toProtoDoc(d)
	}
	return &pb.ListDocumentsResponse{Items: items, Total: total, Start: req.Start, Size: req.Size}, nil
}

func (s *DocumentServiceServer) SaveDocument(ctx context.Context, req *pb.SaveDocumentRequest) (*pb.Document, error) {
	data := decodeData(req.Data)
	doc := &entity.Document{DocumentID: req.DocumentId, Data: data, Locale: req.Locale}
	saved, err := s.uc.Save(ctx, req.ContentTypeSlug, doc, nil, middleware.UserID(ctx))
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoDoc(saved), nil
}

func (s *DocumentServiceServer) PublishDocument(ctx context.Context, req *pb.PublishDocumentRequest) (*pb.Document, error) {
	if err := s.uc.Publish(ctx, req.ContentTypeSlug, req.DocumentId, req.Locale, nil, middleware.UserID(ctx)); err != nil {
		return nil, toGRPCError(err)
	}
	doc, err := s.uc.GetPublished(ctx, req.ContentTypeSlug, req.DocumentId, req.Locale, nil)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoDoc(doc), nil
}

func (s *DocumentServiceServer) UnpublishDocument(ctx context.Context, req *pb.PublishDocumentRequest) (*pb.Document, error) {
	if err := s.uc.Unpublish(ctx, req.ContentTypeSlug, req.DocumentId, req.Locale); err != nil {
		return nil, toGRPCError(err)
	}
	doc, _, err := s.uc.GetForEdit(ctx, req.ContentTypeSlug, req.DocumentId, req.Locale, nil)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoDoc(doc), nil
}

func (s *DocumentServiceServer) DeleteDocument(ctx context.Context, req *pb.DeleteDocumentRequest) (*pb.DeleteDocumentResponse, error) {
	if err := s.uc.Delete(ctx, req.ContentTypeSlug, req.DocumentId, nil); err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.DeleteDocumentResponse{Success: true}, nil
}

func (s *DocumentServiceServer) GetSingleType(ctx context.Context, req *pb.GetSingleTypeRequest) (*pb.Document, error) {
	doc, _, err := s.uc.GetSingleType(ctx, req.ContentTypeSlug, req.Locale, nil)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoDoc(doc), nil
}

func (s *DocumentServiceServer) SaveSingleType(ctx context.Context, req *pb.SaveSingleTypeRequest) (*pb.Document, error) {
	data := decodeData(req.Data)
	saved, err := s.uc.SaveSingleType(ctx, req.ContentTypeSlug, data, req.Locale, nil, middleware.UserID(ctx))
	if err != nil {
		return nil, toGRPCError(err)
	}
	doc, _, err := s.uc.GetSingleType(ctx, req.ContentTypeSlug, saved.Locale, nil)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoDoc(doc), nil
}

func (s *DocumentServiceServer) PublishSingleType(ctx context.Context, req *pb.GetSingleTypeRequest) (*pb.Document, error) {
	if err := s.uc.PublishSingleType(ctx, req.ContentTypeSlug, req.Locale, nil, middleware.UserID(ctx)); err != nil {
		return nil, toGRPCError(err)
	}
	doc, _, err := s.uc.GetSingleType(ctx, req.ContentTypeSlug, req.Locale, nil)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoDoc(doc), nil
}

func (s *DocumentServiceServer) UnpublishSingleType(ctx context.Context, req *pb.GetSingleTypeRequest) (*pb.Document, error) {
	if err := s.uc.UnpublishSingleType(ctx, req.ContentTypeSlug, req.Locale); err != nil {
		return nil, toGRPCError(err)
	}
	doc, _, err := s.uc.GetSingleType(ctx, req.ContentTypeSlug, req.Locale, nil)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoDoc(doc), nil
}

func toProtoDoc(d *entity.Document) *pb.Document {
	pb := &pb.Document{
		DocumentId:  d.DocumentID,
		Version:     string(d.Version),
		Data:        encodeData(d.Data),
		Locale:      d.Locale,
		CreatedAt:   timestamppb.New(d.CreatedAt),
		UpdatedAt:   timestamppb.New(d.UpdatedAt),
		CreatedBy:   d.CreatedBy,
		UpdatedBy:   d.UpdatedBy,
		PublishedBy: d.PublishedBy,
	}
	if !d.PublishedAt.IsZero() {
		pb.PublishedAt = timestamppb.New(d.PublishedAt)
	}
	return pb
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
