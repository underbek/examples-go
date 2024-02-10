package testcontainer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

type TestPostgresSuite struct {
	suite.Suite
	container *PostgresContainer
}

func (s *TestPostgresSuite) SetupSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	var err error
	s.container, err = NewPostgresContainer(ctx)
	s.Require().NoError(err)
}

func (s *TestPostgresSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	s.Require().NoError(s.container.Terminate(ctx))
}

func TestPostgresSuite_Run(t *testing.T) {
	suite.Run(t, new(TestPostgresSuite))
}

func (s *TestPostgresSuite) Test_PostgresConnectionSQLX() {
	db, err := sqlx.Open("postgres", s.container.GetDSN())
	s.Require().NoError(err)
	s.Assert().NoError(db.Close())
}

func (s *TestPostgresSuite) Test_PostgresConnectionPGXPool() {
	poolConfig, err := pgxpool.ParseConfig(s.container.GetDSN())
	s.Require().NoError(err)

	_, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	s.Require().NoError(err)
}

func Test_SetUp(t *testing.T) {
	setup := func() (*PostgresContainer, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
		defer cancel()

		return NewPostgresContainer(ctx)
	}

	teardown := func(container *PostgresContainer) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		return container.Terminate(ctx)
	}

	for i := 0; i < 8; i++ {
		t.Run("setup postgres", func(t *testing.T) {
			t.Parallel()

			for j := 0; j < 8; j++ {
				container, err := setup()
				require.NoError(t, err)

				poolConfig, err := pgxpool.ParseConfig(container.GetDSN())
				require.NoError(t, err)

				db, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
				require.NoError(t, err)

				err = db.Ping(context.Background())
				require.NoError(t, err)

				err = teardown(container)
				require.NoError(t, err)
			}
		})
	}

}
