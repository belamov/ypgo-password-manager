package storage

import "fmt"

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
