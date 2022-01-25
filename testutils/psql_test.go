package testutils

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type TestSuite struct {
	suite.Suite
	container *PostgreSQLContainer
	db *sql.DB
}

func (s *TestSuite) SetupSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer ctxCancel()

	c, err := NewPostgreSQLContainer(ctx)
	s.Require().NoError(err)

	s.container = c

	dsn := c.GetDSN()

	s.db, err = sql.Open("postgres", dsn)
	s.Require().NoError(err)

	err = s.db.Ping()
	s.Require().NoError(err)

	err = migrate(s.db)
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
	s.Run("Ping", func() {
		err := s.db.Ping()
		s.Require().NoError(err)
	})
}
