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
	ErrConfigMissing         = NewInternalError(errors.New("Configuration missing."))
	ErrInvalidPassword       = NewInternalError(errors.New("Invalid password."))
	ErrInvalidSessionToken   = NewInternalError(errors.New("Invalid session token."))
	ErrInvalidUsername       = NewInternalError(errors.New("Invalid username."))
	ErrRuntimeNotFullyInited = NewInternalError(errors.New("Runtime not fully inited."))

	// External error
	ErrInvalidAccount             = NewExternalError(errors.New("Login.InvalidAccount"))
	ErrAccountExists              = NewExternalError(errors.New("Account.Exists"))
	ErrRegisterNotAllowed         = NewExternalError(errors.New("Register.NotAllowed"))
	ErrUsernameNotMail            = NewExternalError(errors.New("Account.NotAMail"))
	ErrWeakPassword               = NewExternalError(errors.New("Account.WeakPassword"))
	ErrSCMPlatformNotFound        = NewExternalError(errors.New("SCM.PlatformNotFound"))
	ErrRepositoryNotFound         = NewExternalError(errors.New("SCM.RepositoryNotFound"))
	ErrSCMPlatformNotSupported    = NewExternalError(errors.New("SCM.PlatformNotSupported"))
	ErrInvalidSCMPlatformID       = NewExternalError(errors.New("SCM.InvalidSCMPlatformID"))
	ErrInvalidRepositoryID        = NewExternalError(errors.New("SCM.InvalidRepositoryID"))
	ErrRepositoryCIAlreadyEnabled = NewExternalError(errors.New("SCM.RepositoryCIAlreadyEnabled"))
	ErrInvalidApprovalID          = NewExternalError(errors.New("SCM.InvalidApprovalID"))
	ErrUnauthenticated            = NewExternalError(errors.New("Auth.Unauthenticated"))

	ErrEndpointMissing = NewExternalError(errors.New("No avaliable endpoint."))
)
