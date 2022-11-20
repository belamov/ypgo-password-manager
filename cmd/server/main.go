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
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
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

	enableTLS := os.Getenv("enable_tls")
	port := os.Getenv("port")

	ctx := context.Background()
	usersRepo, err := storage.NewUserRepository(ctx, os.Getenv("dsn"))
	if err != nil {
		log.Fatal().Err(err).Msg("cant init user repo")
	}

	jwtManager := services.NewJWTManager(os.Getenv("secret_key"), tokenDuration)
	authService := services.NewAuthService(usersRepo)

	address := fmt.Sprintf("0.0.0.0:%s", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}

	err = runGRPCServer(authService, jwtManager, enableTLS == "true", listener)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}

func runGRPCServer(
	authService *services.AuthService,
	jwtManager *services.JWTManager,
	enableTLS bool,
	listener net.Listener,
) error {
	interceptor := proto.NewAuthInterceptor(jwtManager)
	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	}

	if enableTLS {
		tlsCredentials, err := loadTLSCredentials()
		if err != nil {
			return fmt.Errorf("cannot load TLS credentials: %w", err)
		}

		serverOptions = append(serverOptions, grpc.Creds(tlsCredentials))
	}

	authGRCP := grpc2.NewAuthServerService(authService, jwtManager)

	grpcServer := grpc.NewServer(serverOptions...)

	pb.RegisterAuthServer(grpcServer, authGRCP)
	reflection.Register(grpcServer)

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
