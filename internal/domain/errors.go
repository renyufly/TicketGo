package domain

import "errors"

var (
	ErrInvalid         = errors.New("invalid input")
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrForbidden       = errors.New("forbidden")
	ErrNotFound        = errors.New("not found")
	ErrConflict        = errors.New("conflict")
	ErrOutOfStock      = errors.New("out of stock")
	ErrActivityClosed  = errors.New("activity is not active")
)

type Error struct {
	Kind    error
	Message string
	Cause   error
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.Kind }

func New(kind error, message string, cause error) error {
	return &Error{Kind: kind, Message: message, Cause: cause}
}
