package engine

import (
	"github.com/underbek/examples-go/encryption/config"
)

type singleConnection struct {
	connection *Connection
	dsn        string
	path       string
	key        string
	token      string
}

func NewSingleConnection(cfg config.VaultENV, token, key, path string) (ConnectionPool, error) {
	return &singleConnection{
		dsn:   cfg.DSN,
		path:  path,
		key:   key,
		token: token,
	}, nil
}

func (s *singleConnection) Acquire() (*Connection, error) {
	conn, err := newConnection(s.dsn, s.token, s.key, s.path)
	if err != nil {
		return nil, err
	}
	s.connection = conn

	return s.connection, nil
}

func (s *singleConnection) Release(conn *Connection) {
	closeConnection(conn)
}

func (s *singleConnection) Close() {
	closeConnection(s.connection)
}
