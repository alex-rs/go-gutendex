package gutendex

import (
	"errors"
	"fmt"
)

// ErrorKind enumerates categories of errors returned by this package.
type ErrorKind int

const (
	// ErrNetwork indicates a network error during request/response.
	ErrNetwork ErrorKind = iota + 1
	// ErrNotFound indicates the requested resource was not found.
	ErrNotFound
	// ErrRateLimited indicates the request was rate limited (HTTP 429).
	ErrRateLimited
	// ErrServer indicates a server side error (5xx).
	ErrServer
)

// Error provides structured error information for operations.
type Error struct {
	Op   string
	Kind ErrorKind
	Err  error
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Op, e.Err)
	}
	return e.Op
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error { return e.Err }

// Is reports whether the target matches the receiver based on Kind.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	if t.Kind != 0 && e.Kind != t.Kind {
		return false
	}
	return true
}

var errNotFound = &Error{Kind: ErrNotFound}

// IsNotFound reports whether err represents a not-found error.
func IsNotFound(err error) bool {
	return errors.Is(err, errNotFound)
}
