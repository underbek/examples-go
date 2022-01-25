package handlers

import (
	"fmt"
	"github.com/AndreyAndreevich/examples-go/integration_tests/domain"
	"github.com/AndreyAndreevich/examples-go/integration_tests/handlers/mocks"
	"github.com/AndreyAndreevich/examples-go/integration_tests/logic"
	"github.com/AndreyAndreevich/examples-go/integration_tests/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCreateUser(t *testing.T) {
	s := storage.New()
	l := logic.New(s)
	userHandler := NewUserHandler(l)

	fReq, err := os.Open("../fixtures/create_user_request.json")

	request := httptest.NewRequest(http.MethodPost, "/user", fReq)

	w := httptest.NewRecorder()
	userHandler.ServeHTTP(w, request)
	res := w.Result()

	assert.Equal(t, res.StatusCode, http.StatusOK)

	fRes, err := os.ReadFile("../fixtures/create_user_response.json")
	assert.NoError(t, err)
	defer res.Body.Close()
	resBody, _ := ioutil.ReadAll(res.Body)
	assert.JSONEq(t, string(fRes), string(resBody))
}

func TestCreateUserWithMock(t *testing.T) {
	l := &mocks.Logic{}

	l.On("CreateUser", "Tony Stark", mock.Anything).Return(domain.User{
		Id:      10,
		Name:    "Tony Stark",
		Balance: 1_000_000_000,
	}, nil)

	userHandler := NewUserHandler(l)

	fReq, err := os.Open("../fixtures/create_user_request.json")

	request := httptest.NewRequest(http.MethodPost, "/user", fReq)

	w := httptest.NewRecorder()
	userHandler.ServeHTTP(w, request)
	res := w.Result()

	assert.Equal(t, res.StatusCode, http.StatusOK)

	fRes, err := os.ReadFile("../fixtures/create_user_response.json")
	assert.NoError(t, err)
	defer res.Body.Close()
	resBody, _ := ioutil.ReadAll(res.Body)
	assert.JSONEq(t, string(fRes), string(resBody))
}

func TestWithCleanup(t *testing.T) {
	t.Cleanup(func() {
		fmt.Println("Cancel")
	})

	fmt.Println("Done")

	//foo()
}

// go test
