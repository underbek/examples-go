package engine

import (
	"context"

	"github.com/hashicorp/go-kms-wrapping/wrappers/transit/v2"
	"github.com/underbek/examples-go/encryption/config"
	gokitErrors "github.com/underbek/examples-go/errors"
)

type Connection struct {
	wrapper *transit.Wrapper
}

type ConnectionPool interface {
	Close()
	Acquire() (*Connection, error)
	Release(conn *Connection)
}

type connectionPool struct {
	connections     chan *Connection
	connectionSlice []*Connection
}

func NewConnectionPool(cfg config.VaultENV, token, key, path string) (ConnectionPool, error) {
	connections := make(chan *Connection, cfg.PoolSize)
	connectionSlice := make([]*Connection, 0, cfg.PoolSize)

	for i := uint(0); i < cfg.PoolSize; i++ {
		conn, err := newConnection(cfg.DSN, token, key, path)
		if err != nil {
			return nil, err
		}

		connectionSlice = append(connectionSlice, conn)
		connections <- conn
	}

	return &connectionPool{
		connections:     connections,
		connectionSlice: connectionSlice,
	}, nil
}

func (p *connectionPool) Acquire() (*Connection, error) {
	return <-p.connections, nil
}

func (p *connectionPool) Release(conn *Connection) {
	p.connections <- conn
}

func (p *connectionPool) Close() {
	for _, conn := range p.connectionSlice {
		closeConnection(conn)
	}
}

func newConnection(dsn, token, key, path string) (*Connection, error) {
	wrapper := transit.NewWrapper()
	_, err := wrapper.SetConfig(
		context.Background(),
		transit.WithAddress(dsn),
		transit.WithToken(token),
		transit.WithMountPath(path),
		transit.WithKeyName(key),
	)
	if err != nil {
		return nil, gokitErrors.Wrap(err, gokitErrors.TypeInternal, "transit wrapper")
	}

	return &Connection{wrapper}, nil
}

func closeConnection(conn *Connection) {
	if conn == nil {
		return
	}

	conn.wrapper.GetClient().Close()
}
