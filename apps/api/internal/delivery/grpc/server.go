package grpcdelivery

import (
	"google.golang.org/grpc"

	"project-abyssoftime-cms-v2/api/internal/delivery/grpc/interceptor"
	pb "project-abyssoftime-cms-v2/api/proto/cms/v1"
)

func NewServer(
	authUC authUseCase,
	ctUC contentTypeUseCase,
	docUC documentUseCase,
	mediaUC mediaUseCase,
) *grpc.Server {
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.AuthUnaryInterceptor()),
	)
	pb.RegisterAuthServiceServer(srv, NewAuthServiceServer(authUC))
	pb.RegisterContentTypeServiceServer(srv, NewContentTypeServiceServer(ctUC))
	pb.RegisterDocumentServiceServer(srv, NewDocumentServiceServer(docUC))
	pb.RegisterMediaServiceServer(srv, NewMediaServiceServer(mediaUC))
	return srv
}
