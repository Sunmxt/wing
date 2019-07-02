package common

import (
	"errors"
)

type InternalError struct {
	err error
}
type ExternalError struct {
	err error
}

func (e InternalError) Error() string { return e.err.Error() }
func (e ExternalError) Error() string { return e.err.Error() }

func NewInternalError(err error) InternalError {
	return InternalError{err: err}
}

func NewExternalError(err error) ExternalError {
	return ExternalError{err: err}
}

var (
	// Internal error
	ErrConfigMissing       = NewInternalError(errors.New("Configuration missing."))
	ErrInvalidPassword     = NewInternalError(errors.New("Invalid password."))
	ErrInvalidSessionToken = NewInternalError(errors.New("Invalid session token."))
	ErrInvalidUsername     = NewInternalError(errors.New("Invalid username."))

	// External error
	ErrInvalidAccount = NewExternalError(errors.New("Login.InvalidAccount"))
)
