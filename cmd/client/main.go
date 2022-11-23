package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/belamov/ypgo-password-manager/internal/app/models"
	"github.com/belamov/ypgo-password-manager/internal/app/proto"
	"github.com/belamov/ypgo-password-manager/internal/app/steps"
	"github.com/belamov/ypgo-password-manager/pb"
	"github.com/manifoldco/promptui"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"os"
)

func main() {
	ctx := context.Background()
	enableTLS := os.Getenv("ENABLE_TLS") == "true"
	address := os.Getenv("SERVER_ADDRESS")

	address = "127.0.0.1:9000"
	enableTLS = false

	authClientConn, err := NewGRPCConn(ctx, address, enableTLS, "")
	if err != nil {
		log.Fatal().Err(err).Msg("cant init client conn for auth")
	}

	authStep, err := steps.NewAuthStep(authClientConn)
	if err != nil {
		log.Fatal().Err(err).Msg("cant init auth step")
	}

	accessToken, err := authStep.Authorize(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("cant authorize")
	}

	fmt.Println("successfully authenticated!")
	err = authClientConn.Close()
	if err != nil {
		log.Fatal().Err(err).Msg("choose action prompt failed")
	}

	authorizedClienConn, err := NewGRPCConn(ctx, address, enableTLS, accessToken)
	if err != nil {
		log.Fatal().Err(err).Msg("cant init client conn")
	}

	client := pb.NewSecretsClient(authorizedClienConn)

	chooseAction(ctx, client)
}

func chooseAction(ctx context.Context, client pb.SecretsClient) {
	prompt := promptui.Select{
		Label: "What would you like to do?",
		Items: []string{"Get secret", "Add secret"},
	}
	idx, _, err := prompt.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("choose action prompt failed")
	}
	if idx == 0 {
		getSecret(ctx, client)
	}
	if idx == 1 {
		addSecret(ctx, client)
	}
	chooseAction(ctx, client)
}

func getSecret(ctx context.Context, client pb.SecretsClient) {
	secretType := getSecretType()
	secretName := getSecretName()

	req := &pb.GetSecretRequest{Name: secretName}

	switch secretType {
	case models.SecretTypePassword:
		resp, err := client.GetPassword(ctx, req)
		if err != nil {
			fmt.Println("Cant get your secret!")
			log.Fatal().Err(err).Msg("cant get password from server")
		}
		fmt.Printf("Login: %s\n", resp.Login)
		fmt.Printf("Password: %s\n", resp.Password)
		return
	case models.SecretTypeCard:
		resp, err := client.GetCard(ctx, req)
		if err != nil {
			fmt.Println("Cant get your secret!")
			log.Fatal().Err(err).Msg("cant get card from server")
		}
		fmt.Printf("Number: %s\n", resp.Number)
		fmt.Printf("Holder name: %s\n", resp.HolderName)
		fmt.Printf("Date: %s\n", resp.Date)
		fmt.Printf("CCV: %s\n", resp.Ccv)
	case models.SecretTypeText:
		resp, err := client.GetText(ctx, req)
		if err != nil {
			fmt.Println("Cant get your secret!")
			log.Fatal().Err(err).Msg("cant get text from server")
		}
		fmt.Printf("Text: %s\n", resp.Text)
	}
	fmt.Println(secretType, secretName)
}

func addSecret(ctx context.Context, client pb.SecretsClient) {
	secretType := getSecretType()
	secretName := getSecretName()

	switch secretType {
	case models.SecretTypePassword:
		req := &pb.SavePasswordRequest{
			Name:     secretName,
			Login:    getValueFromUser("Enter login"),
			Password: getValueFromUser("Enter password"),
		}
		_, err := client.SavePassword(ctx, req)
		if err != nil {
			fmt.Println("Cant save your secret!")
			log.Fatal().Err(err).Msg("cant save password on server")
		}
		fmt.Println("Password Saved!")
		return
	case models.SecretTypeCard:
		req := &pb.SaveCardRequest{
			CardName:   secretName,
			Number:     getValueFromUser("Enter card number"),
			HolderName: getValueFromUser("Enter card holder name"),
			Date:       getValueFromUser("Enter card date"),
			Ccv:        getValueFromUser("Enter card ccv"),
		}
		_, err := client.SaveCard(ctx, req)
		if err != nil {
			fmt.Println("Cant save your secret!")
			log.Fatal().Err(err).Msg("cant save card on server")
		}
		fmt.Println("Card Saved!")
		return
	case models.SecretTypeText:
		req := &pb.SaveTextRequest{
			Name: secretName,
			Text: getValueFromUser("Enter text"),
		}
		_, err := client.SaveText(ctx, req)
		if err != nil {
			fmt.Println("Cant save your secret!")
			log.Fatal().Err(err).Msg("cant save text on server")
		}
		fmt.Println("Text Saved!")
		return
	}
}

func getValueFromUser(label string) string {
	prompt := promptui.Prompt{
		Label: label,
	}

	value, err := prompt.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("cant get login for password secret")
	}

	return value
}

func getSecretName() string {
	prompt := promptui.Prompt{
		Label: "Enter secret name",
	}

	result, err := prompt.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("choose secret type prompt failed")
	}

	return result
}

func getSecretType() models.SecretType {
	secretTypes := []models.SecretType{
		models.SecretTypePassword,
		models.SecretTypeCard,
		models.SecretTypeText,
	}
	prompt := promptui.Select{
		Label: "Select type of secret",
		Items: secretTypes,
	}
	idx, _, err := prompt.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("choose secret type prompt failed")
	}

	return secretTypes[idx]
}

func NewGRPCConn(ctx context.Context, address string, enableTLS bool, accessToken string) (*grpc.ClientConn, error) {
	transportOption := grpc.WithTransportCredentials(insecure.NewCredentials())

	if enableTLS {
		tlsCredentials, err := loadTLSCredentials()
		if err != nil {
			return nil, err
		}

		transportOption = grpc.WithTransportCredentials(tlsCredentials)
	}

	if accessToken != "" {
		clientAuthInterceptor := proto.NewClientAuthInterceptor(accessToken)
		return getClientConn(
			ctx,
			address,
			transportOption,
			grpc.WithUnaryInterceptor(clientAuthInterceptor.Unary()),
			grpc.WithUnaryInterceptor(clientAuthInterceptor.Unary()),
		)
	}
	return getClientConn(ctx, address, transportOption)
}

func getClientConn(ctx context.Context, address string, transportOptions ...grpc.DialOption) (*grpc.ClientConn, error) {
	conn, err := grpc.DialContext(
		ctx,
		address,
		transportOptions...,
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	clientCert, err := tls.LoadX509KeyPair("cert/client-cert.pem", "cert/client-key.pem")
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
	}

	return credentials.NewTLS(config), nil
}
