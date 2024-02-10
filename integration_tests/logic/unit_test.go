package logic

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/underbek/examples-go/integration_tests/domain"
)

func TestCreatUser(t *testing.T) {
	storageMock := &storageMock{}
	l := New(storageMock)

	storageMock.On("GetUser", mock.Anything, 10).Return(domain.User{
		ID:      10,
		Name:    "Tony Stark",
		Balance: decimal.NewFromInt(1_000_000_000),
	}, nil)

	user, err := l.GetUser(context.Background(), 10)
	assert.NoError(t, err)

	assert.Equal(t, 10, user.ID)
	assert.Equal(t, "Tony Stark", user.Name)
	assert.Equal(t, "1000000000", user.Balance.String())
}
