package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	grpc2 "github.com/belamov/ypgo-password-manager/internal/app/grpc"
	"github.com/belamov/ypgo-password-manager/internal/app/proto"
	"github.com/belamov/ypgo-password-manager/internal/app/services"
	"github.com/belamov/ypgo-password-manager/internal/app/storage"
	"github.com/belamov/ypgo-password-manager/pb"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"net"
	"os"
	"time"
)

var (
	buildVersion = "N/A" //nolint:gochecknoglobals
	buildDate    = "N/A" //nolint:gochecknoglobals
	buildCommit  = "N/A" //nolint:gochecknoglobals

	tokenDuration = 15 * time.Minute
)

const (
	serverCertFile   = "cert/server-cert.pem"
	serverKeyFile    = "cert/server-key.pem"
	clientCACertFile = "cert/ca-cert.pem"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()

	log.Info().Msgf("Build version: %s\n", buildVersion)
	log.Info().Msgf("Build date: %s\n", buildDate)
	log.Info().Msgf("Build commit: %s\n", buildCommit)

	enableTLS := os.Getenv("enable_tls") == "true"
	port := os.Getenv("port")
	dsn := os.Getenv("dsn")
	secretkey := os.Getenv("secret_key")

	enableTLS = false
	port = "9000"
	dsn = "postgres://postgres:postgres@db:5432/praktikum?sslmode=disable"
	secretkey = "secret key secret key secret key"

	ctx := context.Background()

	err := storage.RunMigrations(dsn)
	if err != nil {
		log.Fatal().Err(err).Msg("cant run migrations")
	}

	usersRepo, err := storage.NewUserRepository(ctx, dsn)
	if err != nil {
		log.Fatal().Err(err).Msg("cant init user repo")
	}

	secretsRepo, err := storage.NewSecretsRepository(ctx, dsn)
	if err != nil {
		log.Fatal().Err(err).Msg("cant init secrets repo")
	}

	jwtManager := services.NewJWTManager(secretkey, tokenDuration)
	authService := services.NewAuthService(usersRepo)

	crypto := &services.GCMAESCryptographer{
		Random: &services.TrulyRandomGenerator{},
		Key:    []byte(secretkey),
	}
	secretsService := services.NewSecretsService(secretsRepo, crypto)

	address := fmt.Sprintf("0.0.0.0:%s", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
	err = runGRPCServer(authService, secretsService, jwtManager, enableTLS, listener)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}

func runGRPCServer(
	authService *services.AuthService,
	secretsService *services.SecretsManager,
	jwtManager *services.JWTManager,
	enableTLS bool,
	listener net.Listener,
) error {
	interceptor := proto.NewAuthInterceptor(jwtManager)
	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(),
			interceptor.Unary(),
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_recovery.StreamServerInterceptor(),
			interceptor.Stream(),
		)),
	}

	if enableTLS {
		tlsCredentials, err := loadTLSCredentials()
		if err != nil {
			return fmt.Errorf("cannot load TLS credentials: %w", err)
		}

		serverOptions = append(serverOptions, grpc.Creds(tlsCredentials))
	}

	grpcServer := grpc.NewServer(serverOptions...)

	authGRCP := grpc2.NewAuthServerService(authService, jwtManager)
	secretsGRPC := grpc2.NewSecretsServerService(secretsService, jwtManager)

	pb.RegisterAuthServer(grpcServer, authGRCP)
	pb.RegisterSecretsServer(grpcServer, secretsGRPC)

	log.Info().Msgf("Start GRPC server at %s, TLS = %t", listener.Addr().String(), enableTLS)
	return grpcServer.Serve(listener)
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed client's certificate
	pemClientCA, err := ioutil.ReadFile(clientCACertFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, fmt.Errorf("failed to add client CA's certificate")
	}

	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair(serverCertFile, serverKeyFile)
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(config), nil
}
