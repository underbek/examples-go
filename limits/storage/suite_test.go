package storage

import (
	"context"
	"database/sql"
	"github.com/underbek/examples-go/limits/testutils"
	"github.com/underbek/examples-go/logger"
	"testing"
	"time"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/suite"
	"github.com/underbek/examples-go/limits/migrations"
	goKitMigrations "github.com/underbek/examples-go/migrate"
	goKitPgx "github.com/underbek/examples-go/storage/pgx"
	"github.com/underbek/examples-go/testcontainers"
)

type TestSuite struct {
	suite.Suite
	postgresContainer *testcontainer.PostgresContainer
	db                goKitPgx.Storage
	storage           *Storage
}

func TestSuiteStorage_Run(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer ctxCancel()

	l, err := logger.New(true)
	s.Require().NoError(err)

	s.postgresContainer, err = testcontainer.NewPostgresContainer(ctx)
	s.Require().NoError(err)

	time.Sleep(time.Second * 10)

	err = goKitMigrations.Run(
		s.postgresContainer.GetDSN(),
		goKitMigrations.WithFs(migrations.Migrations),
		goKitMigrations.WithDriver("pgx"),
		goKitMigrations.WithLogger(l),
	)
	s.Require().NoError(err)

	db, err := sql.Open("pgx", s.postgresContainer.GetDSN())
	s.Require().NoError(err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("pgx"),
		testfixtures.FS(testutils.Fixtures),
		testfixtures.Directory("fixtures/storage"),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
	s.Require().NoError(db.Close())

	s.db, err = goKitPgx.New(
		context.Background(),
		goKitPgx.Config{DSN: s.postgresContainer.GetDSN(), Timeout: time.Minute},
		goKitPgx.WithLogger(l),
	)
	s.Require().NoError(err)

	s.storage = New(l, s.db, 100)
}

func (s *TestSuite) TearDownSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	s.db.Close()
	s.Require().NoError(s.postgresContainer.Terminate(ctx))
}

func (s *TestSuite) SetupTest() {
	s.SetupSubTest()
}

func (s *TestSuite) SetupSubTest() {
	db, err := sql.Open("pgx", s.postgresContainer.GetDSN())
	s.Require().NoError(err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("pgx"),
		testfixtures.FS(testutils.Fixtures),
		testfixtures.Directory("fixtures/storage"),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
	s.Require().NoError(db.Close())
}
