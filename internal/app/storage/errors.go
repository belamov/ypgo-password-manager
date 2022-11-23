package storage

import (
	"fmt"
	"github.com/belamov/ypgo-password-manager/internal/app/models"
)

type NotUniqueUsernameError struct {
	Err      error
	Username string
}

func (err *NotUniqueUsernameError) Error() string {
	return fmt.Sprintf("username already exists: %s", err.Username)
}

func (err *NotUniqueUsernameError) Unwrap() error {
	return err.Err
}

func NewNotUniqueUsernameError(username string, err error) error {
	return &NotUniqueUsernameError{
		Err:      err,
		Username: username,
	}
}

type UserNotFoundError struct {
	Err      error
	Username string
}

func (err *UserNotFoundError) Error() string {
	return fmt.Sprintf("user not found: login = %s", err.Username)
}

func (err *UserNotFoundError) Unwrap() error {
	return err.Err
}

func NewUserNotFoundError(username string, err error) error {
	return &UserNotFoundError{
		Err:      err,
		Username: username,
	}
}

type NotUniqueSecretError struct {
	Err      error
	Metadata models.SecretMetadata
}

func (err *NotUniqueSecretError) Error() string {
	return fmt.Sprintf("metadato of secret is not unique (%q)", err.Metadata)
}

func (err *NotUniqueSecretError) Unwrap() error {
	return err.Err
}

func NewNotUniqueSecret(metadata models.SecretMetadata, err error) error {
	return &NotUniqueSecretError{
		Err:      err,
		Metadata: metadata,
	}
}

type SecretNotFoundError struct {
	Err      error
	Metadata models.SecretMetadata
}

func (err *SecretNotFoundError) Error() string {
	return fmt.Sprintf("secret not found (%q)", err.Metadata)
}

func (err *SecretNotFoundError) Unwrap() error {
	return err.Err
}

func NewSecretNotFoundError(metadata models.SecretMetadata, err error) error {
	return &SecretNotFoundError{
		Err:      err,
		Metadata: metadata,
	}
}
