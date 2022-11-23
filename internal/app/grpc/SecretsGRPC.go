package grpc

import (
	"context"
	"github.com/belamov/ypgo-password-manager/internal/app/models"
	"github.com/belamov/ypgo-password-manager/internal/app/services"
	"github.com/belamov/ypgo-password-manager/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type SecretsGRPC struct {
	pb.UnimplementedSecretsServer
	secretsService services.SecretsManagerInterface
	jwtManager     *services.JWTManager
}

func (s *SecretsGRPC) getUserId(ctx context.Context) (int, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return 0, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	accessToken := values[0]
	claims, err := s.jwtManager.Verify(accessToken)
	if err != nil {
		return 0, status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	}

	return claims.Id, nil
}
func (s *SecretsGRPC) SavePassword(ctx context.Context, request *pb.SavePasswordRequest) (*pb.Empty, error) {
	userID, err := s.getUserId(ctx)
	if err != nil {
		return nil, err
	}

	secretPassword := &models.PasswordSecret{
		Login:    request.GetLogin(),
		Password: request.GetPassword(),
	}
	secretMetadata := models.SecretMetadata{
		Name:   request.GetName(),
		Type:   models.SecretTypePassword,
		UserID: userID,
	}

	secretData, err := secretPassword.ToBinary()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot encode secret: %v", err)
	}

	err = s.secretsService.SaveSecret(secretData, secretMetadata)

	return &pb.Empty{}, err
}

func (s *SecretsGRPC) GetPassword(ctx context.Context, request *pb.GetSecretRequest) (*pb.PasswordResponse, error) {
	userID, err := s.getUserId(ctx)
	if err != nil {
		return nil, err
	}

	secretMetadata := models.SecretMetadata{
		Name:   request.GetName(),
		Type:   models.SecretTypePassword,
		UserID: userID,
	}

	secret, err := s.secretsService.GetPassword(secretMetadata)
	if err != nil {
		return nil, err
	}

	response := &pb.PasswordResponse{
		Login:    secret.Login,
		Password: secret.Password,
	}

	return response, nil
}

func (s *SecretsGRPC) SaveCard(ctx context.Context, request *pb.SaveCardRequest) (*pb.Empty, error) {
	userID, err := s.getUserId(ctx)
	if err != nil {
		return nil, err
	}

	secret := &models.CardSecret{
		Number:     request.GetNumber(),
		HolderName: request.GetHolderName(),
		CCV:        request.GetCcv(),
		Date:       request.GetDate(),
	}
	secretMetadata := models.SecretMetadata{
		Name:   request.GetCardName(),
		Type:   models.SecretTypeCard,
		UserID: userID,
	}

	secretData, err := secret.ToBinary()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot encode secret: %v", err)
	}

	err = s.secretsService.SaveSecret(secretData, secretMetadata)

	return &pb.Empty{}, err
}

func (s *SecretsGRPC) GetCard(ctx context.Context, request *pb.GetSecretRequest) (*pb.CardResponse, error) {
	userID, err := s.getUserId(ctx)
	if err != nil {
		return nil, err
	}

	secretMetadata := models.SecretMetadata{
		Name:   request.GetName(),
		Type:   models.SecretTypeCard,
		UserID: userID,
	}

	secret, err := s.secretsService.GetCard(secretMetadata)
	if err != nil {
		return nil, err
	}

	response := &pb.CardResponse{
		Number:     secret.Number,
		HolderName: secret.HolderName,
		Date:       secret.Date,
		Ccv:        secret.CCV,
	}

	return response, nil
}

func (s *SecretsGRPC) SaveText(ctx context.Context, request *pb.SaveTextRequest) (*pb.Empty, error) {
	userID, err := s.getUserId(ctx)
	if err != nil {
		return nil, err
	}

	secret := &models.TextSecret{
		Text: request.GetText(),
	}
	secretMetadata := models.SecretMetadata{
		Name:   request.GetName(),
		Type:   models.SecretTypeText,
		UserID: userID,
	}

	secretData, err := secret.ToBinary()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot encode secret: %v", err)
	}

	err = s.secretsService.SaveSecret(secretData, secretMetadata)

	return &pb.Empty{}, err
}

func (s *SecretsGRPC) GetText(ctx context.Context, request *pb.GetSecretRequest) (*pb.TextResponse, error) {
	userID, err := s.getUserId(ctx)
	if err != nil {
		return nil, err
	}

	secretMetadata := models.SecretMetadata{
		Name:   request.GetName(),
		Type:   models.SecretTypeText,
		UserID: userID,
	}

	secret, err := s.secretsService.GetText(secretMetadata)
	if err != nil {
		return nil, err
	}

	response := &pb.TextResponse{Text: secret.Text}

	return response, nil
}

func NewSecretsServerService(service services.SecretsManagerInterface, manager *services.JWTManager) *SecretsGRPC {
	return &SecretsGRPC{secretsService: service, jwtManager: manager}
}
