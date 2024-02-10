package testcontainer

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/go-kms-wrapping/wrappers/transit/v2"
	"github.com/stretchr/testify/suite"
	"gopkg.in/resty.v1"
)

type TestVaultSuite struct {
	suite.Suite
	container *VaultContainer
}

func (s *TestVaultSuite) SetupSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	var err error
	s.container, err = NewVaultContainer(ctx)
	s.Require().NoError(err)
}

func (s *TestVaultSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	s.Require().NoError(s.container.Terminate(ctx))
}

func TestVaultSuite_Run(t *testing.T) {
	suite.Run(t, new(TestVaultSuite))
}

func (s *TestVaultSuite) Test_TransitEngine() {
	// turn on transit engine -> 204 No Content
	resp, err := resty.R().
		SetHeader("X-Vault-Token", s.container.GetToken()).
		SetBody(`{"type":"transit"}`).
		Post(s.container.GetDSN() + "/v1/sys/mounts/transit")

	s.Require().NoError(err)
	s.Require().Equal(http.StatusNoContent, resp.StatusCode())

	w := transit.NewWrapper()

	_, err = w.SetConfig(
		context.Background(),
		transit.WithAddress(s.container.GetDSN()),
		transit.WithToken(s.container.GetToken()),
		transit.WithMountPath("transit"),
		transit.WithKeyName("example"),
	)
	s.Require().NoError(err)

	input := "test_value"
	encryptedValue, err := w.GetClient().Encrypt([]byte(input))
	s.Require().NoError(err)

	decryptedValue, err := w.GetClient().Decrypt(encryptedValue)
	s.Require().NoError(err)

	s.Require().Equal(input, string(decryptedValue))
}
