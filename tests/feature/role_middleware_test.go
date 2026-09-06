package feature

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/models"
	"github.com/ridhoauliama97/pos-server/database/seeders"
	"github.com/ridhoauliama97/pos-server/tests"
)

type RoleMiddlewareTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestRoleMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(RoleMiddlewareTestSuite))
}

func (s *RoleMiddlewareTestSuite) SetupTest() {
	s.RefreshDatabase()
	s.Seed(&seeders.DatabaseSeeder{})

	s.createUser("Owner", "owner@pos.local", models.RoleOwner)
	s.createUser("Kasir", "kasir@pos.local", models.RoleKasir)
}

func (s *RoleMiddlewareTestSuite) createUser(name, email, role string) {
	password, err := facades.Hash().Make("password123")
	s.NoError(err)

	if err := facades.Orm().Query().Create(&models.User{
		Name:     name,
		Email:    email,
		Password: password,
		Role:     role,
		IsActive: true,
	}); err != nil {
		s.Fail("failed to create user: %s", err.Error())
	}
}

func (s *RoleMiddlewareTestSuite) login(email string) string {
	resp, err := s.Http(s.T()).WithHeader("Content-Type", "application/json").Post("/api/v1/auth/login",
		strings.NewReader(`{"email":"`+email+`","password":"password123"}`),
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

func (s *RoleMiddlewareTestSuite) TestKasirCannotAccessUserManagement() {
	token := s.login("kasir@pos.local")

	resp, err := s.Http(s.T()).WithToken(token).Get("/api/v1/users")
	s.NoError(err)
	resp.AssertForbidden()

	resp, err = s.Http(s.T()).WithToken(token).WithHeader("Content-Type", "application/json").Post("/api/v1/users",
		strings.NewReader(`{"name":"X","email":"x@pos.local","password":"secret123","role":"kasir"}`),
	)
	s.NoError(err)
	resp.AssertForbidden()
}

func (s *RoleMiddlewareTestSuite) TestAdminCannotCreateOwner() {
	token := s.login("admin@pos.local")

	resp, err := s.Http(s.T()).WithToken(token).WithHeader("Content-Type", "application/json").Post("/api/v1/users",
		strings.NewReader(`{"name":"New Owner","email":"owner2@pos.local","password":"secret123","role":"owner"}`),
	)
	s.NoError(err)
	resp.AssertForbidden()

	resp, err = s.Http(s.T()).WithToken(token).WithHeader("Content-Type", "application/json").Post("/api/v1/users",
		strings.NewReader(`{"name":"New Kasir","email":"kasir2@pos.local","password":"secret123","role":"kasir"}`),
	)
	s.NoError(err)
	resp.AssertCreated()
}

func (s *RoleMiddlewareTestSuite) TestOwnerCanAccessAllEndpoints() {
	token := s.login("owner@pos.local")

	resp, err := s.Http(s.T()).WithToken(token).Get("/api/v1/users")
	s.NoError(err)
	resp.AssertOk()

	resp, err = s.Http(s.T()).WithToken(token).WithHeader("Content-Type", "application/json").Post("/api/v1/users",
		strings.NewReader(`{"name":"New Kasir","email":"kasir3@pos.local","password":"secret123","role":"kasir"}`),
	)
	s.NoError(err)
	resp.AssertCreated()

	resp, err = s.Http(s.T()).WithToken(token).Delete("/api/v1/users/1", nil)
	s.NoError(err)
	resp.AssertOk()
}
