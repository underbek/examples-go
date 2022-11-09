package testcontainer

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/suite"
)

type TestRedisSuite struct {
	suite.Suite
	container *RedisContainer
}

func (s *TestRedisSuite) SetupSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	var err error
	s.container, err = NewRedisContainer(ctx)
	s.Require().NoError(err)
}

func (s *TestRedisSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	s.container.Terminate(ctx)
}

func TestSuiteRedis_Run(t *testing.T) {
	suite.Run(t, new(TestRedisSuite))
}

func (s *TestRedisSuite) Test_RedisConn() {
	cli := redis.NewClient(&redis.Options{
		Addr:     s.container.GetHost(),
		Password: s.container.GetPassword(),
	})

	cmd := cli.Ping(context.Background())

	s.Require().NoError(cmd.Err())
	s.Assert().NoError(cli.Close())
}
