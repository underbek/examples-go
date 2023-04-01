package errors

import (
	"net/http"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
)

func TestError_New(t *testing.T) {
	testWantErr := &Error{
		errType: TypeUnknown,
		message: "some error",
	}

	testGotErr := New(TypeUnknown, "some error")

	assert.Equal(t, testGotErr.errType, testWantErr.errType)
	assert.Equal(t, testGotErr.Error(), testWantErr.Error())
	assert.Equal(t, testGotErr.Error(), "[Unknown] some error")
}

func TestError_Wrap(t *testing.T) {
	tests := []struct {
		name               string
		testWantErrType    Type
		testWantErrMessage string
		testGotErr         error
	}{
		{
			name:               "wrap basic error",
			testWantErrType:    TypeUnknown,
			testWantErrMessage: "TestError_Wrap error",
			testGotErr:         Wrap(errors.New("some error"), TypeUnknown, "TestError_Wrap error"),
		},
		{
			name:               "wrapf basic error",
			testWantErrType:    TypeUnknown,
			testWantErrMessage: "TestError_Wrapf error",
			testGotErr: Wrapf(
				errors.New("some error"),
				TypeUnknown,
				"%s %s",
				"TestError_Wrapf",
				"error",
			),
		},
		{
			name:               "wrap lib error",
			testWantErrType:    TypeNotFound,
			testWantErrMessage: "some error",
			testGotErr: Wrap(
				New(TypeNotFound, "some error"),
				TypeUnknown,
				"TestError_Wrap error",
			),
		},
		{
			name:               "wrapf lib error",
			testWantErrType:    TypeNotFound,
			testWantErrMessage: "some error",
			testGotErr: Wrapf(
				New(TypeNotFound, "some error"),
				TypeUnknown,
				"%s %s",
				"TestError_Wrapf",
				"error",
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errType, message := parseError(tt.testGotErr)
			assert.Equal(t, tt.testWantErrType, errType)
			assert.Equal(t, tt.testWantErrMessage, message)
		})
	}
}

func TestError_Error(t *testing.T) {
	tests := []struct {
		name           string
		testGotErr     error
		testWantErrStr string
	}{
		{
			name: "simple error",
			testGotErr: &Error{
				errType: TypeNotFound,
				message: "some error",
			},
			testWantErrStr: "[NotFound] some error",
		},
		{
			name: "with internal error",
			testGotErr: &Error{
				errType: TypeNotFound,
				message: "some error",
				err:     errors.New("Internal error"),
			},
			testWantErrStr: "[NotFound] some error: Internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.testGotErr.Error(), tt.testWantErrStr)
		})
	}
}

func TestError_ParseError(t *testing.T) {
	type testCase struct {
		name       string
		err        error
		statusCode int
		grpcCode   codes.Code
		errMsg     string
	}

	testCases := []testCase{
		{
			name:       "error type",
			err:        errors.New("some error"),
			statusCode: http.StatusInternalServerError,
			grpcCode:   codes.Internal,
			errMsg:     "application error",
		},
		{
			name: "Error type invalid request",
			err: &Error{
				errType: TypeInvalidRequest,
				message: "some error",
			},
			statusCode: http.StatusBadRequest,
			grpcCode:   codes.InvalidArgument,
			errMsg:     "some error",
		},
		{
			name: "Error type TypeUnknown error",
			err: &Error{
				errType: TypeUnknown,
				message: "some error",
			},
			statusCode: http.StatusInternalServerError,
			grpcCode:   codes.Internal,
			errMsg:     "some error",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			gotStatusCode, gotHttpErr := ParseHttpError(tt.err)
			gotGRPCCode, gotGRPCErr := ParseGRPCError(tt.err)

			assert.Equal(t, gotStatusCode, tt.statusCode)
			assert.Equal(t, gotGRPCCode, tt.grpcCode)
			assert.Equal(t, gotHttpErr, tt.errMsg)
			assert.Equal(t, gotGRPCErr, tt.errMsg)
		})
	}
}
