// Code generated by mockery v2.11.0. DO NOT EDIT.

package logic

import (
	context "context"

	domain "github.com/AndreyAndreevich/examples-go/integration_tests/domain"
	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// storageMock is an autogenerated mock type for the storage type
type storageMock struct {
	mock.Mock
}

// CreatUser provides a mock function with given fields: ctx, user
func (_m *storageMock) CreatUser(ctx context.Context, user domain.User) (domain.User, error) {
	ret := _m.Called(ctx, user)

	var r0 domain.User
	if rf, ok := ret.Get(0).(func(context.Context, domain.User) domain.User); ok {
		r0 = rf(ctx, user)
	} else {
		r0 = ret.Get(0).(domain.User)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, domain.User) error); ok {
		r1 = rf(ctx, user)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUser provides a mock function with given fields: ctx, id
func (_m *storageMock) GetUser(ctx context.Context, id int) (domain.User, error) {
	ret := _m.Called(ctx, id)

	var r0 domain.User
	if rf, ok := ret.Get(0).(func(context.Context, int) domain.User); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(domain.User)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateUser provides a mock function with given fields: ctx, user
func (_m *storageMock) UpdateUser(ctx context.Context, user domain.User) (domain.User, error) {
	ret := _m.Called(ctx, user)

	var r0 domain.User
	if rf, ok := ret.Get(0).(func(context.Context, domain.User) domain.User); ok {
		r0 = rf(ctx, user)
	} else {
		r0 = ret.Get(0).(domain.User)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, domain.User) error); ok {
		r1 = rf(ctx, user)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// newStorageMock creates a new instance of storageMock. It also registers a cleanup function to assert the mocks expectations.
func newStorageMock(t testing.TB) *storageMock {
	mock := &storageMock{}

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
