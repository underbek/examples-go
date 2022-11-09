package errors

import (
	"net/http"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
)

// Error is the type that implements the error interface.
type Error struct {
	Type Type
	Err  error
}

// New method allows for making errors.
func New(errType Type, msg string) *Error {
	return &Error{
		Type: errType,
		Err:  errors.New(msg),
	}
}

// Errorf method allows for making errors and the format specifier.
func Errorf(errType Type, format string, args ...interface{}) *Error {
	return &Error{
		Type: errType,
		Err:  errors.Errorf(format, args...),
	}
}

// Wrap method allows for wrapping errors.
func Wrap(errType Type, err error, msg string) *Error {
	return &Error{
		Type: wrapErrorType(errType, err),
		Err:  errors.Wrap(err, msg),
	}
}

// Wrapf method allows for wrapping errors and the format specifier.
func Wrapf(errType Type, err error, format string, args ...interface{}) *Error {
	return &Error{
		Type: wrapErrorType(errType, err),
		Err:  errors.Wrapf(err, format, args...),
	}
}

func wrapErrorType(errType Type, err error) Type {
	cErr, ok := err.(*Error)
	if !ok {
		return errType
	}

	return cErr.Type
}

// Error method do string from error type.
func (e *Error) Error() string {
	return e.Err.Error()
}

// Type defines the type of error this is.
type Type = string

// Types of errors.
const (
	// InvalidRequest - invalid request from user.
	InvalidRequest Type = "invalid request error"
	// NotFound - item does not exist.
	NotFound Type = "not found"
	// Unauthorized request.
	Unauthorized Type = "unauthorized request"
	// Database - error from database.
	Database Type = "database error"
	// Internal - error or inconsistency.
	Internal Type = "internal error"
	// External - error or inconsistency.
	External Type = "external error"
	// Other - unclassified error. This value is not printed in the error message.
	Other Type = "other error"
)

func ErrorType(err error) Type {
	cErr, ok := err.(*Error)
	if !ok {
		return Other
	}

	return cErr.Type
}

// ParseError parses errors from error and Error types.
func ParseError(err error) (int, string) {
	cErr, ok := err.(*Error)
	if !ok {
		return http.StatusInternalServerError, Internal
	}

	return cErr.ParseErrorType()
}

// ParseErrorType parses Error by types.
func (e *Error) ParseErrorType() (int, string) {
	switch e.Type {
	case InvalidRequest:
		return http.StatusBadRequest, InvalidRequest
	case NotFound:
		return http.StatusNotFound, NotFound
	case Unauthorized:
		return http.StatusUnauthorized, Unauthorized
	case Database:
		return http.StatusInternalServerError, Database
	case External:
		return http.StatusInternalServerError, External
	}

	return http.StatusInternalServerError, Other
}

// ParseGRPCError parses errors from error and Error types.
func ParseGRPCError(err error) (codes.Code, string) {
	cErr, ok := err.(*Error)
	if !ok {
		return codes.Internal, Internal
	}

	return cErr.ParseGRPCCode()
}

// ParseGRPCCode parses Error by types.
func (e *Error) ParseGRPCCode() (codes.Code, string) {
	switch e.Type {
	case InvalidRequest:
		return codes.InvalidArgument, InvalidRequest
	case NotFound:
		return codes.NotFound, NotFound
	case Unauthorized:
		return codes.Unauthenticated, Unauthorized
	case Database:
		return codes.Internal, Database
	case External:
		return codes.Unavailable, External
	}

	return codes.Internal, Other
}
