package storage

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/underbek/examples-go/integration_tests/domain"
)

type storage struct {
	db *sqlx.DB
}

var (
	ErrAlreadyExists          = errors.New("already exists")
	ErrNotFound               = errors.New("not found")
	ErrIncorrectQueryResponse = errors.New("incorrect query response")
)

func New(dsn string) (*storage, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	return &storage{
		db: db,
	}, nil
}

func (s *storage) CreatUser(ctx context.Context, user domain.User) (domain.User, error) {
	query := `INSERT INTO users (name, balance) 
				VALUES (:name, :balance) 
				RETURNING id, name, balance, created_at, updated_at;`

	res, err := s.db.NamedQueryContext(ctx, query, &user)
	if err != nil {
		return domain.User{}, err
	}
	defer res.Close()

	if !res.Next() {
		return domain.User{}, ErrIncorrectQueryResponse
	}
	var resUser domain.User
	if err := res.StructScan(&resUser); err != nil {
		return domain.User{}, err
	}

	return resUser, nil
}

func (s *storage) GetUser(ctx context.Context, id int) (domain.User, error) {
	query := `SELECT id, name, balance, created_at, updated_at FROM users WHERE id = $1`
	var user domain.User
	err := s.db.GetContext(ctx, &user, query, id)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (s *storage) UpdateUser(_ context.Context, _ domain.User) (domain.User, error) {
	return domain.User{}, nil
}
