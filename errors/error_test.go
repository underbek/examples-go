package errors_test

import (
	"net/http"
	"testing"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"

	"github.com/magiconair/properties/assert"
	errors_lib "github.com/underbek/examples-go/errors"
)

func TestError_New(t *testing.T) {
	testWantErr := &errors_lib.Error{
		Type: errors_lib.Other,
		Err:  errors.New("some error"),
	}

	testGotErr := errors_lib.New(errors_lib.Other, "some error")

	assert.Equal(t, testGotErr.Type, testWantErr.Type)
	assert.Equal(t, testGotErr.Err.Error(), testWantErr.Err.Error())
}

func TestError_Wrap(t *testing.T) {
	tests := []struct {
		name        string
		testWantErr *errors_lib.Error
		testGotErr  *errors_lib.Error
	}{
		{
			name: "wrap basic error",
			testWantErr: &errors_lib.Error{
				Type: errors_lib.Other,
				Err:  errors.Wrap(errors.New("some error"), "TestError_Wrap error"),
			},

			testGotErr: errors_lib.Wrap(errors_lib.Other, errors.New("some error"), "TestError_Wrap error"),
		},
		{
			name: "wrapf basic error",
			testWantErr: &errors_lib.Error{
				Type: errors_lib.Other,
				Err:  errors.Wrapf(errors.New("some error"), "%s %s", "TestError_Wrapf", "error"),
			},

			testGotErr: errors_lib.Wrapf(
				errors_lib.Other,
				errors.New("some error"),
				"%s %s",
				"TestError_Wrapf",
				"error",
			),
		},
		{
			name: "wrap lib error",
			testWantErr: &errors_lib.Error{
				Type: errors_lib.NotFound,
				Err:  errors.Wrap(errors.New("some error"), "TestError_Wrap error"),
			},

			testGotErr: errors_lib.Wrap(
				errors_lib.Other,
				errors_lib.New(errors_lib.NotFound, "some error"),
				"TestError_Wrap error",
			),
		},
		{
			name: "wrapf lib error",
			testWantErr: &errors_lib.Error{
				Type: errors_lib.NotFound,
				Err:  errors.Wrapf(errors.New("some error"), "%s %s", "TestError_Wrapf", "error"),
			},

			testGotErr: errors_lib.Wrapf(
				errors_lib.Other,
				errors_lib.New(errors_lib.NotFound, "some error"),
				"%s %s",
				"TestError_Wrapf",
				"error",
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.testGotErr.Type, tt.testWantErr.Type)
			assert.Equal(t, tt.testGotErr.Err.Error(), tt.testWantErr.Err.Error())
		})
	}
}

func TestError_Error(t *testing.T) {
	testErr := errors_lib.Error{
		Type: errors_lib.Other,
		Err:  errors.New("some error"),
	}

	testWantErrStr := "some error"
	testGotErrStr := testErr.Error()

	assert.Equal(t, testGotErrStr, testWantErrStr)
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
			errMsg:     errors_lib.Internal,
		},
		{
			name: "errors_lib.Error type invalid request",
			err: &errors_lib.Error{
				Type: errors_lib.InvalidRequest,
				Err:  errors.New("some error"),
			},
			statusCode: http.StatusBadRequest,
			grpcCode:   codes.InvalidArgument,
			errMsg:     errors_lib.InvalidRequest,
		},
		{
			name: "errors_lib.Error type other error",
			err: &errors_lib.Error{
				Type: errors_lib.Other,
				Err:  errors.New("some error"),
			},
			statusCode: http.StatusInternalServerError,
			grpcCode:   codes.Internal,
			errMsg:     errors_lib.Other,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			gotStatusCode, gotHttpErr := errors_lib.ParseError(tt.err)
			gotGRPCCode, gotGRPCErr := errors_lib.ParseGRPCError(tt.err)

			assert.Equal(t, gotStatusCode, tt.statusCode)
			assert.Equal(t, gotGRPCCode, tt.grpcCode)
			assert.Equal(t, gotHttpErr, tt.errMsg)
			assert.Equal(t, gotGRPCErr, tt.errMsg)
		})
	}
}

func TestError_parseErrorType(t *testing.T) {
	type testCase struct {
		name       string
		err        interface{}
		statusCode int
		grpcCode   codes.Code
		errMsg     string
	}

	testCases := []testCase{
		{
			name: "invalid request error",
			err: &errors_lib.Error{
				Type: errors_lib.InvalidRequest,
				Err:  errors.New("some invalid request error"),
			},
			statusCode: http.StatusBadRequest,
			grpcCode:   codes.InvalidArgument,
			errMsg:     errors_lib.InvalidRequest,
		},
		{
			name: "other error",
			err: &errors_lib.Error{
				Type: errors_lib.Other,
				Err:  errors.New("some other error"),
			},
			statusCode: http.StatusInternalServerError,
			grpcCode:   codes.Internal,
			errMsg:     errors_lib.Other,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			gotStatusCode, gotHttpErr := tt.err.(*errors_lib.Error).ParseErrorType()
			gotGRPCCode, gotGRPCErr := tt.err.(*errors_lib.Error).ParseGRPCCode()

			assert.Equal(t, gotStatusCode, tt.statusCode)
			assert.Equal(t, gotGRPCCode, tt.grpcCode)
			assert.Equal(t, gotHttpErr, tt.errMsg)
			assert.Equal(t, gotGRPCErr, tt.errMsg)
		})
	}
}
