package errors

import "fmt"

// Error is the type that implements the error interface.
type Error struct {
	errType Type
	message string
	err     error
}

// New method allows for making errors.
func New(errType Type, msg string) *Error {
	return &Error{
		errType: errType,
		message: msg,
	}
}

// Errorf method allows for making errors and the format specifier.
func Errorf(errType Type, format string, args ...interface{}) *Error {
	return &Error{
		errType: errType,
		message: fmt.Sprintf(format, args...),
	}
}

// Wrap method allows for wrapping errors.
func Wrap(err error, errType Type, msg string) *Error {
	return &Error{
		errType: errType,
		message: msg,
		err:     err,
	}
}

// Wrapf method allows for wrapping errors and the format specifier.
func Wrapf(err error, errType Type, format string, args ...interface{}) *Error {
	return &Error{
		errType: errType,
		message: fmt.Sprintf(format, args...),
		err:     err,
	}
}

// Error method do string from error type.
func (e *Error) Error() string {
	if e.err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.errType, e.message, e.err)
	}

	return fmt.Sprintf("[%s] %s", e.errType, e.message)
}

// Unwrap returns the result of calling the Unwrap method on err, if err implements
// Unwrap. Otherwise, Unwrap returns nil.
func (e *Error) Unwrap() error {
	return e.err
}
