package usererror

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidCredential = errors.New("invalid credential")
	ErrInvalidToken = errors.New("invalid token")
)