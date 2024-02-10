package pgx

import (
	"context"
	testcontainer "github.com/underbek/examples-go/testcontainers"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	postgresContainer *testcontainer.PostgresContainer
}

func TestSuite_Run(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer ctxCancel()

	var err error
	s.postgresContainer, err = testcontainer.NewPostgresContainer(ctx)
	s.Require().NoError(err)
}

func (s *TestSuite) TearDownSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	s.Require().NoError(s.postgresContainer.Terminate(ctx))
}
