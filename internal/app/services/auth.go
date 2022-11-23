package services

import (
	"fmt"
	"github.com/belamov/ypgo-password-manager/internal/app/models"
	"github.com/belamov/ypgo-password-manager/internal/app/storage"
	"golang.org/x/crypto/bcrypt"
)

type Auth interface {
	Register(username string, password string) (models.User, error)
	Login(username string, password string) (models.User, error)
}

type AuthService struct {
	userStorage storage.UserStorage
}

func NewAuthService(userStorage storage.UserStorage) *AuthService {
	return &AuthService{userStorage: userStorage}
}

func (a *AuthService) Register(username string, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("cannot hash password: %w", err)
	}

	registeredUser, err := a.userStorage.CreateNew(username, string(hashedPassword))
	if err != nil {
		return nil, fmt.Errorf("cannot save user: %w", err)
	}

	return registeredUser, nil
}

func (a *AuthService) Login(username string, password string) (*models.User, error) {
	user, err := a.userStorage.Find(username)
	if err != nil {
		return nil, fmt.Errorf("cannot find user: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("incorrect password: %w", err)
	}

	return user, nil
}
