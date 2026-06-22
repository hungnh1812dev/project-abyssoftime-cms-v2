package grpcdelivery

import (
	"context"
	"strings"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pb "project-abyssoftime-cms-v2/api/proto/cms/v1"
)

type authUseCase interface {
	Register(ctx context.Context, email, password, displayName string) (*entity.User, error)
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	SetupStatus(ctx context.Context) (bool, error)
}

type AuthServiceServer struct {
	pb.UnimplementedAuthServiceServer
	usecase authUseCase
}

func NewAuthServiceServer(usecase authUseCase) *AuthServiceServer {
	return &AuthServiceServer{usecase: usecase}
}

func (server *AuthServiceServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	access, refresh, err := server.usecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.LoginResponse{AccessToken: access, RefreshToken: refresh}, nil
}

func (server *AuthServiceServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	displayName := req.Email
	if idx := strings.Index(req.Email, "@"); idx > 0 {
		displayName = req.Email[:idx]
	}
	user, err := server.usecase.Register(ctx, req.Email, req.Password, displayName)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.RegisterResponse{Id: user.DocumentID, Email: user.Email, Role: string(user.Role)}, nil
}

func (server *AuthServiceServer) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	access, _, err := server.usecase.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.RefreshResponse{AccessToken: access}, nil
}

func (server *AuthServiceServer) SetupStatus(ctx context.Context, _ *pb.SetupStatusRequest) (*pb.SetupStatusResponse, error) {
	hasAdmin, err := server.usecase.SetupStatus(ctx)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.SetupStatusResponse{HasAdmin: hasAdmin}, nil
}
