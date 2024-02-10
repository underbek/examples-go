package pgx

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *TestSuite) TestSuiteTimeout() {
	st, err := New(context.Background(), Config{
		DSN:     s.postgresContainer.GetDSN(),
		Timeout: time.Second * 2,
	})
	s.Require().NoError(err)

	_, err = st.Exec(context.Background(), "SELECT pg_sleep(10)")
	s.Require().Error(err)
	s.Require().ErrorContains(err, "canceling statement due to statement timeout (SQLSTATE 57014)")

	_, err = st.Exec(context.Background(), "SELECT pg_sleep(1)")
	s.Require().NoError(err)
}

func (s *TestSuite) TestSuiteCustomTxTimeout() {
	st, err := New(context.Background(), Config{
		DSN:     s.postgresContainer.GetDSN(),
		Timeout: time.Second * 2,
	})
	s.Require().NoError(err)

	tx, err := st.Begin(context.Background(), &pgx.TxOptions{})
	s.Require().NoError(err)

	_, err = tx.Exec(context.Background(), "SELECT pg_sleep(10)")
	s.Require().Error(err)
	s.Require().ErrorContains(err, "canceling statement due to statement timeout (SQLSTATE 57014)")

	tx, err = st.Begin(context.Background(), &pgx.TxOptions{})
	s.Require().NoError(err)

	_, err = tx.Exec(context.Background(), "SELECT pg_sleep(1)")
	s.Require().NoError(err)

	err = tx.Rollback(context.Background())
	s.Require().NoError(err)
}
