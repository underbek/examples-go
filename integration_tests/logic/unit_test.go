package logic

import (
	"github.com/AndreyAndreevich/examples-go/integration_tests/domain"
	"github.com/AndreyAndreevich/examples-go/integration_tests/logic/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreatUser(t *testing.T) {
	storageMock := &mocks.Storage{}
	l := New(storageMock)

	storageMock.On("GetUser", 10).Return(domain.User{
		Id:      10,
		Name:    "Tony Stark",
		Balance: 1_000_000_000,
	}, nil)

	user, err := l.GetUser(10)
	assert.NoError(t, err)

	assert.Equal(t, 10, user.Id)
	assert.Equal(t, "Tony Stark", user.Name)
	assert.Equal(t, float64(1_000_000_000), user.Balance)
}
