package feature

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/models"
	"github.com/ridhoauliama97/pos-server/database/seeders"
	"github.com/ridhoauliama97/pos-server/tests"
)

type AuthTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}

func (s *AuthTestSuite) SetupTest() {
	s.RefreshDatabase()
	s.Seed(&seeders.DatabaseSeeder{})
}

func (s *AuthTestSuite) login(email, password string) string {
	resp, err := s.Http(s.T()).WithHeader("Content-Type", "application/json").Post("/api/v1/auth/login",
		strings.NewReader(`{"email":"`+email+`","password":"`+password+`"}`),
	)
	s.NoError(err)
	resp.AssertOk()

	body, err := resp.Json()
	s.NoError(err)
	token, ok := body["token"].(string)
	s.True(ok)
	s.NotEmpty(token)

	return token
}

func (s *AuthTestSuite) TestLoginValidCredentials() {
	s.login("admin@pos.local", "password123")
}

func (s *AuthTestSuite) TestLoginInvalidCredentials() {
	resp, err := s.Http(s.T()).WithHeader("Content-Type", "application/json").Post("/api/v1/auth/login",
		strings.NewReader(`{"email":"admin@pos.local","password":"wrong-password"}`),
	)
	s.NoError(err)
	resp.AssertUnauthorized()
}

func (s *AuthTestSuite) TestLoginInactiveUser() {
	var user models.User
	s.NoError(facades.Orm().Query().Where("email", "admin@pos.local").FirstOrFail(&user))
	_, err := facades.Orm().Query().Model(&user).Update("is_active", false)
	s.NoError(err)

	resp, err := s.Http(s.T()).WithHeader("Content-Type", "application/json").Post("/api/v1/auth/login",
		strings.NewReader(`{"email":"admin@pos.local","password":"password123"}`),
	)
	s.NoError(err)
	resp.AssertForbidden()
}

func (s *AuthTestSuite) TestLogoutInvalidatesToken() {
	token := s.login("admin@pos.local", "password123")

	resp, err := s.Http(s.T()).WithToken(token).Post("/api/v1/auth/logout", nil)
	s.NoError(err)
	resp.AssertOk()

	resp, err = s.Http(s.T()).WithToken(token).Post("/api/v1/auth/refresh", nil)
	s.NoError(err)
	resp.AssertUnauthorized()
}

func (s *AuthTestSuite) TestRefreshIssuesNewToken() {
	// The preceding logout test disables the token it issued for the same user in
	// ITS second. JWT iat/exp are second-granularity with no random nonce, so
	// logging in during that same second would reproduce the identical (now
	// disabled) token string. Sleep first so login lands in a new second.
	time.Sleep(1100 * time.Millisecond)

	original := s.login("admin@pos.local", "password123")

	time.Sleep(1100 * time.Millisecond)

	resp, err := s.Http(s.T()).WithToken(original).Post("/api/v1/auth/refresh", nil)
	if err != nil {
		s.T().Logf("REFRESH_FLAKE refresh err=%v", err)
	}
	s.NoError(err)
	resp.AssertOk()

	body, err := resp.Json()
	s.NoError(err)
	token, ok := body["token"].(string)
	s.True(ok)
	s.NotEmpty(token)
	s.NotEqual(token, original)
}
