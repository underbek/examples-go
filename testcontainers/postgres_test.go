package testcontainer

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	container *PostgresContainer
}

func (s *TestSuite) SetupSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	var err error
	s.container, err = NewPostgresContainer(ctx)
	s.Require().NoError(err)
}

func (s *TestSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	s.container.Terminate(ctx)
}

func TestSuite_Run(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) Test_PostgresConnectionSQLX() {
	db, err := sqlx.Open("postgres", s.container.GetDSN())
	s.Require().NoError(err)
	s.Assert().NoError(db.Close())
}

func (s *TestSuite) Test_PostgresConnectionPGXPool() {
	poolConfig, err := pgxpool.ParseConfig(s.container.GetDSN())
	s.Require().NoError(err)

	_, err = pgxpool.ConnectConfig(context.Background(), poolConfig)
	s.Require().NoError(err)
}
