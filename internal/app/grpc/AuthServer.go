package grpc

import (
	"context"
	"github.com/belamov/ypgo-password-manager/internal/app/services"
	"github.com/belamov/ypgo-password-manager/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthGRCP struct {
	pb.UnimplementedAuthServer
	authService *services.AuthService
	jwtManager  *services.JWTManager
}

func NewAuthServerService(service *services.AuthService, manager *services.JWTManager) *AuthGRCP {
	return &AuthGRCP{authService: service, jwtManager: manager}
}

func (a *AuthGRCP) Login(_ context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := a.authService.Login(request.Username, request.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot login user", err)
	}

	token, err := a.jwtManager.Generate(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate access token")
	}

	res := &pb.LoginResponse{AccessToken: token}
	return res, nil
}

func (a *AuthGRCP) Register(_ context.Context, request *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user, err := a.authService.Register(request.Username, request.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot register user", err)
	}

	token, err := a.jwtManager.Generate(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate access token")
	}

	res := &pb.RegisterResponse{AccessToken: token}
	return res, nil
}
