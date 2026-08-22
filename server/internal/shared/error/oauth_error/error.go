package oautherror

import "errors"

var (
	ErrInvalidState = errors.New("invalid state")
	ErrInvalidProvider = errors.New("invalid provider")
)