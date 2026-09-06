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

type AdvancedReportTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestAdvancedReportTestSuite(t *testing.T) {
	suite.Run(t, new(AdvancedReportTestSuite))
}

func (s *AdvancedReportTestSuite) SetupTest() {
	s.RefreshDatabase()
	s.Seed(&seeders.DatabaseSeeder{})
}

func (s *AdvancedReportTestSuite) seedKasir(name, email string) int64 {
	password, err := facades.Hash().Make("password123")
	s.NoError(err)
	user := models.User{Name: name, Email: email, Password: password, Role: models.RoleKasir, IsActive: true}
	s.NoError(facades.Orm().Query().Create(&user))

	return int64(user.ID)
}

func (s *AdvancedReportTestSuite) login(email string) string {
	resp, err := s.Http(s.T()).WithHeader("Content-Type", "application/json").Post("/api/v1/auth/login",
		strings.NewReader(`{"email":"`+email+`","password":"password123"}`),
	)
	s.NoError(err)
	resp.AssertOk()
	body, _ := resp.Json()

	return body["token"].(string)
}

func (s *AdvancedReportTestSuite) seedVariant() int64 {
	category := &models.Category{Name: "Foods"}
	s.NoError(facades.Orm().Query().Create(category))
	product := models.Product{CategoryId: int64(category.ID), Name: "Nasi Goreng", IsVariant: true, BasePrice: 15000, IsActive: true}
	s.NoError(facades.Orm().Query().Create(&product))
	variant := models.ProductVariant{ProductId: int64(product.ID), Name: "Nasi Goreng", Price: 15000}
	s.NoError(facades.Orm().Query().Create(&variant))
	s.NoError(facades.Orm().Query().Create(&models.Stock{OutletId: 1, ProductVariantId: int64(variant.ID), Quantity: 50}))

	return int64(variant.ID)
}

func (s *AdvancedReportTestSuite) createTransaction(token, uuid string, variantId int64, quantity int) {
	resp, err := s.Http(s.T()).WithToken(token).WithHeader("Content-Type", "application/json").Post("/api/v1/transactions",
		strings.NewReader(fmt.Sprintf(`{
			"client_uuid": "%s", "outlet_id": 1, "tax_percent": 10,
			"items": [{"product_variant_id": %d, "quantity": %d}],
			"payments": [{"payment_method_id": 1, "amount": 100000}]
		}`, uuid, variantId, quantity)),
	)
	s.NoError(err)
	resp.AssertCreated()
}

func (s *AdvancedReportTestSuite) TestByKasirGroupsSalesPerCashier() {
	s.seedKasir("Kasir Satu", "kasir1@pos.local")
	s.seedKasir("Kasir Dua", "kasir2@pos.local")
	variantId := s.seedVariant()
	kasir1Token := s.login("kasir1@pos.local")
	kasir2Token := s.login("kasir2@pos.local")
	adminToken := s.login("admin@pos.local")

	s.createTransaction(kasir1Token, "advk-1", variantId, 2)
	s.createTransaction(kasir1Token, "advk-2", variantId, 1)
	s.createTransaction(kasir2Token, "advk-3", variantId, 3)

	resp, err := s.Http(s.T()).WithToken(adminToken).Get("/api/v1/reports/kasir")
	s.NoError(err)
	resp.AssertOk()
	body, _ := resp.Json()
	items := body["data"].(map[string]any)["items"].([]any)
	s.Len(items, 2)

	kasir1 := items[0].(map[string]any)
	s.Equal("Kasir Satu", kasir1["cashier_name"])
	s.Equal(2.0, kasir1["transaction_count"])
	s.Equal(45000.0, kasir1["subtotal"])

	// kasir sees only own group
	resp, err = s.Http(s.T()).WithToken(kasir1Token).Get("/api/v1/reports/kasir")
	s.NoError(err)
	resp.AssertOk()
	body, _ = resp.Json()
	items = body["data"].(map[string]any)["items"].([]any)
	s.Len(items, 1)
	s.Equal("Kasir Satu", items[0].(map[string]any)["cashier_name"])
}

func (s *AdvancedReportTestSuite) TestShiftReportScopesByRole() {
	_ = s.seedKasir("Kasir Satu", "kasir1@pos.local")
	kasir2Id := s.seedKasir("Kasir Dua", "kasir2@pos.local")
	variantId := s.seedVariant()
	kasir1Token := s.login("kasir1@pos.local")
	kasir2Token := s.login("kasir2@pos.local")
	adminToken := s.login("admin@pos.local")

	s.createTransaction(kasir1Token, "advs-1", variantId, 2)
	s.createTransaction(kasir2Token, "advs-2", variantId, 4)

	// kasir sees own shift only even when passing kasir_id of another
	resp, err := s.Http(s.T()).WithToken(kasir1Token).Get("/api/v1/reports/shift?kasir_id=999")
	s.NoError(err)
	resp.AssertOk()
	body, _ := resp.Json()
	data := body["data"].(map[string]any)
	s.Equal("Kasir Satu", data["cashier_name"])
	s.Equal(1.0, data["per_day"].([]any)[0].(map[string]any)["transaction_count"])

	// admin must pass kasir_id
	resp, err = s.Http(s.T()).WithToken(adminToken).Get("/api/v1/reports/shift")
	s.NoError(err)
	resp.AssertStatus(422)

	resp, err = s.Http(s.T()).WithToken(adminToken).Get(fmt.Sprintf("/api/v1/reports/shift?kasir_id=%d", kasir2Id))
	s.NoError(err)
	resp.AssertOk()
	body, _ = resp.Json()
	data = body["data"].(map[string]any)
	s.Equal("Kasir Dua", data["cashier_name"])
	day := data["per_day"].([]any)[0].(map[string]any)
	s.Equal(1.0, day["transaction_count"])
	s.Equal(4.0, day["item_quantity"])
}

func (s *AdvancedReportTestSuite) TestProductsReportFiltersByOutlet() {
	variantId := s.seedVariant()
	token := s.login("admin@pos.local")
	s.createTransaction(token, "advp-1", variantId, 2)

	resp, err := s.Http(s.T()).WithToken(token).Get("/api/v1/reports/products?outlet_id=1")
	s.NoError(err)
	resp.AssertOk()
	body, _ := resp.Json()
	s.Len(body["data"].(map[string]any)["items"].([]any), 1)

	resp, err = s.Http(s.T()).WithToken(token).Get("/api/v1/reports/products?outlet_id=99")
	s.NoError(err)
	resp.AssertOk()
	body, _ = resp.Json()
	s.Len(body["data"].(map[string]any)["items"].([]any), 0)
}
