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
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.AuthUnaryInterceptor()),
	)
	pb.RegisterAuthServiceServer(grpcServer, NewAuthServiceServer(authUC))
	pb.RegisterContentTypeServiceServer(grpcServer, NewContentTypeServiceServer(ctUC))
	pb.RegisterDocumentServiceServer(grpcServer, NewDocumentServiceServer(docUC))
	pb.RegisterMediaServiceServer(grpcServer, NewMediaServiceServer(mediaUC))
	return grpcServer
}
