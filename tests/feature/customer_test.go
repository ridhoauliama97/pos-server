package feature

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/models"
	"github.com/ridhoauliama97/pos-server/database/seeders"
	"github.com/ridhoauliama97/pos-server/tests"
)

type CustomerTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestCustomerTestSuite(t *testing.T) {
	suite.Run(t, new(CustomerTestSuite))
}

func (s *CustomerTestSuite) SetupTest() {
	s.RefreshDatabase()
	s.Seed(&seeders.DatabaseSeeder{})

	password, err := facades.Hash().Make("password123")
	s.NoError(err)
	s.NoError(facades.Orm().Query().Create(&models.User{
		Name:     "Kasir",
		Email:    "kasir@pos.local",
		Password: password,
		Role:     models.RoleKasir,
		IsActive: true,
	}))
}

func (s *CustomerTestSuite) login(email string) string {
	resp, err := s.Http(s.T()).WithHeader("Content-Type", "application/json").Post("/api/v1/auth/login",
		strings.NewReader(`{"email":"`+email+`","password":"password123"}`),
	)
	s.NoError(err)
	resp.AssertOk()
	body, _ := resp.Json()

	return body["token"].(string)
}

func (s *CustomerTestSuite) seedShop() (string, int64) {
	adminToken := s.login("admin@pos.local")

	category := &models.Category{Name: "Foods"}
	s.NoError(facades.Orm().Query().Create(category))
	product := models.Product{CategoryId: int64(category.ID), Name: "Nasi Goreng", IsVariant: true, BasePrice: 15000, IsActive: true}
	s.NoError(facades.Orm().Query().Create(&product))
	variant := models.ProductVariant{ProductId: int64(product.ID), Name: "Nasi Goreng", Price: 15000}
	s.NoError(facades.Orm().Query().Create(&variant))
	s.NoError(facades.Orm().Query().Create(&models.Stock{OutletId: 1, ProductVariantId: int64(variant.ID), Quantity: 50}))

	return adminToken, int64(variant.ID)
}

func (s *CustomerTestSuite) createCustomer(token string) int64 {
	resp, err := s.Http(s.T()).WithToken(token).WithHeader("Content-Type", "application/json").Post("/api/v1/customers",
		strings.NewReader(`{"name":"Budi","phone":"0812","email":"budi@x.com"}`),
	)
	s.NoError(err)
	resp.AssertCreated()
	body, _ := resp.Json()
	data := body["data"].(map[string]any)
	s.NotEmpty(data["member_code"])

	return int64(data["id"].(float64))
}

func (s *CustomerTestSuite) TestCustomerCrudAndMemberCodeGeneration() {
	token := s.login("admin@pos.local")

	id := s.createCustomer(token)

	resp, err := s.Http(s.T()).WithToken(token).Get(fmt.Sprintf("/api/v1/customers/%d", id))
	s.NoError(err)
	resp.AssertOk()
	body, _ := resp.Json()
	s.Equal("Budi", body["data"].(map[string]any)["name"])

	resp, err = s.Http(s.T()).WithToken(token).WithHeader("Content-Type", "application/json").Put(fmt.Sprintf("/api/v1/customers/%d", id),
		strings.NewReader(`{"name":"Budi Updated"}`),
	)
	s.NoError(err)
	resp.AssertOk()

	resp, err = s.Http(s.T()).WithToken(token).Get("/api/v1/customers?q=Budi")
	s.NoError(err)
	resp.AssertOk()
	body, _ = resp.Json()
	s.Len(body["data"].([]any), 1)

	kasirToken := s.login("kasir@pos.local")
	resp, err = s.Http(s.T()).WithToken(kasirToken).Get("/api/v1/customers?q=0812")
	s.NoError(err)
	resp.AssertOk()

	resp, err = s.Http(s.T()).WithToken(kasirToken).Delete(fmt.Sprintf("/api/v1/customers/%d", id), nil)
	s.NoError(err)
	resp.AssertForbidden()

	resp, err = s.Http(s.T()).WithToken(token).Delete(fmt.Sprintf("/api/v1/customers/%d", id), nil)
	s.NoError(err)
	resp.AssertOk()
}

func (s *CustomerTestSuite) TestCustomerHistoryReturnsCompletedTransactions() {
	adminToken, variantId := s.seedShop()
	customerId := s.createCustomer(adminToken)

	createTx := func(uuid string) {
		resp, err := s.Http(s.T()).WithToken(adminToken).WithHeader("Content-Type", "application/json").Post("/api/v1/transactions",
			strings.NewReader(fmt.Sprintf(`{
				"client_uuid": "%s", "outlet_id": 1, "customer_id": %d, "tax_percent": 10,
				"items": [{"product_variant_id": %d, "quantity": 1}],
				"payments": [{"payment_method_id": 1, "amount": 50000}]
			}`, uuid, customerId, variantId)),
		)
		s.NoError(err)
		resp.AssertCreated()
	}
	createTx("cust-tx-1")
	createTx("cust-tx-2")

	resp, err := s.Http(s.T()).WithToken(adminToken).Get(fmt.Sprintf("/api/v1/customers/%d/transactions?limit=1&page=1", customerId))
	s.NoError(err)
	resp.AssertOk()
	body, _ := resp.Json()
	items := body["data"].([]any)
	s.Len(items, 1)
	meta := body["meta"].(map[string]any)
	s.Equal(2.0, meta["total"])
	s.Equal(1.0, meta["limit"])
	item := items[0].(map[string]any)
	s.Len(item["items"].([]any), 1)
	s.Equal("Nasi Goreng", item["items"].([]any)[0].(map[string]any)["product_name_snapshot"])
}
