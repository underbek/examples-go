package errors

import (
	"errors"
	"net/http"

	"google.golang.org/grpc/codes"
)

// ParseError parses errors from error and Error types.
func parseError(err error) (Type, string) {
	errType := TypeUnknown
	message := "application error"
	for ; err != nil; err = errors.Unwrap(err) {
		switch cErr := err.(type) {
		case *Error:
			errType = cErr.errType
			message = cErr.message
		}
	}

	return errType, message
}

func ErrorType(err error) Type {
	errType, _ := parseError(err)

	return errType
}

// ParseHttpError parses errors from error and Error types.
func ParseHttpError(err error) (int, string) {
	errType, message := parseError(err)
	code := http.StatusInternalServerError

	switch errType {
	case TypeInvalidRequest:
		code = http.StatusBadRequest
	case TypeNotFound:
		code = http.StatusNotFound
	case TypeUnauthorized:
		code = http.StatusUnauthorized
	case TypeDatabase:
		code = http.StatusInternalServerError
	case TypeExternal, TypeInternal:
		code = http.StatusInternalServerError
	case TypeNotImplemented:
		code = http.StatusNotImplemented
	}

	return code, message
}

// ParseGRPCError parses errors from error and Error types.
func ParseGRPCError(err error) (codes.Code, string) {
	errType, message := parseError(err)
	code := codes.Internal

	switch errType {
	case TypeInvalidRequest:
		code = codes.InvalidArgument
	case TypeNotFound:
		code = codes.NotFound
	case TypeUnauthorized:
		code = codes.Unauthenticated
	case TypeDatabase:
		code = codes.Internal
	case TypeExternal, TypeInternal:
		code = codes.Unavailable
	case TypeNotImplemented:
		code = codes.Unimplemented
	}

	return code, message
}
