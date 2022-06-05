package storage

import (
	"context"
	"errors"

	"github.com/AndreyAndreevich/examples-go/integration_tests/domain"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type storage struct {
	db *sqlx.DB
}

var (
	AlreadyExistsErr       = errors.New("already exists")
	NotFoundErr            = errors.New("not found")
	IncorrectQueryResponse = errors.New("incorrect query response")
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
		return domain.User{}, IncorrectQueryResponse
	}
	var resUser domain.User
	if err := res.StructScan(&resUser); err != nil {
		return domain.User{}, err
	}

	return resUser, nil
}

func (s *storage) GetUser(ctx context.Context, id int) (domain.User, error) {
	return domain.User{}, nil
}

func (s *storage) UpdateUser(ctx context.Context, user domain.User) (domain.User, error) {
	return domain.User{}, nil
}
