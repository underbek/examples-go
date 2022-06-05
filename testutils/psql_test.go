package testutils

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	container *PostgreSQLContainer
	db        *sql.DB
}

func (s *TestSuite) SetupSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer ctxCancel()

	c, err := NewPostgreSQLContainer(ctx)
	s.Require().NoError(err)

	s.container = c

	s.db, err = sql.Open("postgres", c.GetDSN())
	s.Require().NoError(err)
}

func (s *TestSuite) TearDownSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	s.Require().NoError(s.container.Terminate(ctx))

}

func TestSuite_PostgreSQLStorage(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestPing() {
	s.Require().NoError(s.db.Ping())
}
