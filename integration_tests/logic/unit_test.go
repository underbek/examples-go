package logic

import (
	"context"
	"testing"

	"github.com/AndreyAndreevich/examples-go/integration_tests/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreatUser(t *testing.T) {
	storageMock := &storageMock{}
	l := New(storageMock)

	storageMock.On("GetUser", mock.Anything, 10).Return(domain.User{
		Id:      10,
		Name:    "Tony Stark",
		Balance: decimal.NewFromInt(1_000_000_000),
	}, nil)

	user, err := l.GetUser(context.Background(), 10)
	assert.NoError(t, err)

	assert.Equal(t, 10, user.Id)
	assert.Equal(t, "Tony Stark", user.Name)
	assert.Equal(t, "1000000000", user.Balance.String())
}
