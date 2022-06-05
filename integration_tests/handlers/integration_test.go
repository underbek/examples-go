package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AndreyAndreevich/examples-go/integration_tests/logic"
	"github.com/AndreyAndreevich/examples-go/integration_tests/migrate"
	"github.com/AndreyAndreevich/examples-go/integration_tests/storage"
	"github.com/AndreyAndreevich/examples-go/integration_tests/testentities"
	"github.com/AndreyAndreevich/examples-go/testutils"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	container   *testutils.PostgreSQLContainer
	fixtures    *testutils.FixtureLoader
	userHandler *userHandler
}

func (s *TestSuite) SetupSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer ctxCancel()

	c, err := testutils.NewPostgreSQLContainer(ctx)
	s.Require().NoError(err)

	s.container = c

	s.fixtures = testutils.NewFixtureLoader(s.T(), testentities.Fixtures)

	db, err := sql.Open("postgres", c.GetDSN())
	s.Require().NoError(err)

	err = testutils.Migrate(db, migrate.Migrations)
	s.Require().NoError(err)

	repo, err := storage.New(c.GetDSN())
	s.Require().NoError(err)

	l := logic.New(repo)
	s.userHandler = &userHandler{l}
}

func (s *TestSuite) TearDownSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	s.Require().NoError(s.container.Terminate(ctx))
}

func TestSuite_PostgreSQLStorage(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestCreateUser() {
	fReq := s.fixtures.LoadString("fixtures/create_user_request.json")

	request := httptest.NewRequest(http.MethodPost, "/user", bytes.NewBufferString(fReq))

	w := httptest.NewRecorder()
	s.userHandler.ServeHTTP(w, request)
	res := w.Result()
	defer res.Body.Close()

	s.Require().Equal(res.StatusCode, http.StatusOK)

	var userResponse UserResponse
	err := json.NewDecoder(res.Body).Decode(&userResponse)
	s.Require().NoError(err)

	expected := s.fixtures.LoadTemplate("fixtures/create_user_response.json.temp", map[string]interface{}{
		"id": userResponse.Id,
	})

	testutils.JSONEq(s.T(), expected, userResponse)
}
