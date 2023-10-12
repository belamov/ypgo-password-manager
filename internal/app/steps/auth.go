package steps

import (
	"context"
	"errors"
	"fmt"
	"github.com/belamov/ypgo-password-manager/pb"
	"github.com/manifoldco/promptui"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type AuthStep struct {
	client pb.AuthClient
}

func NewAuthStep(conn *grpc.ClientConn) (*AuthStep, error) {
	client := pb.NewAuthClient(conn)

	return &AuthStep{client: client}, nil
}

func (a *AuthStep) Authorize(ctx context.Context) (string, error) {

	prompt := promptui.Select{
		Label: "Welcome!",
		Items: []string{"Login", "Register"},
	}

	idx, _, err := prompt.Run()
	if err != nil {
		fmt.Printf("prompt failed %v\n", err)
		return "", err
	}

	if idx == 0 {
		return a.login(ctx)
	}

	if idx == 1 {
		return a.register(ctx)
	}

	return "", errors.New("unexpected input")
}

func (a *AuthStep) login(ctx context.Context) (string, error) {
	username, err := getUsername()
	if err != nil {
		return "", fmt.Errorf("cannot get username: %w", err)
	}

	password, err := getPassword()
	if err != nil {
		return "", fmt.Errorf("cannot get password: %w", err)
	}

	request := &pb.LoginRequest{
		Username: username,
		Password: password,
	}
	response, err := a.client.Login(ctx, request)
	if err != nil {
		return "", fmt.Errorf("cannot login: %w", err)
	}

	return response.AccessToken, nil
}

func (a *AuthStep) register(ctx context.Context) (string, error) {
	username, err := getUsername()
	if err != nil {
		return "", fmt.Errorf("cannot get username: %w", err)
	}

	password, err := getPassword()
	if err != nil {
		return "", fmt.Errorf("cannot get password: %w", err)
	}

	request := &pb.RegisterRequest{
		Username: username,
		Password: password,
	}
	response, err := a.client.Register(ctx, request)
	log.Info().Msgf("%q", response)
	log.Info().Msg("resp")
	log.Info().Err(err)
	if err != nil {
		return "", fmt.Errorf("cannot login: %w", err)
	}

	return response.AccessToken, nil
}

func getUsername() (string, error) {
	prompt := promptui.Prompt{
		Label: "Enter your username",
	}
	return prompt.Run()
}

func getPassword() (string, error) {
	prompt := promptui.Prompt{
		Label: "Enter your username",
		Mask:  '*',
	}
	return prompt.Run()
}
