package feature

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goravel/framework/support/carbon"
	"github.com/stretchr/testify/suite"

	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/models"
	"github.com/ridhoauliama97/pos-server/database/seeders"
	"github.com/ridhoauliama97/pos-server/tests"
)

type ReportTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestReportTestSuite(t *testing.T) {
	suite.Run(t, new(ReportTestSuite))
}

func (s *ReportTestSuite) SetupTest() {
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

func (s *ReportTestSuite) login(email string) string {
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

func (s *ReportTestSuite) seedProduct() int64 {
	category := &models.Category{Name: "Foods"}
	s.NoError(facades.Orm().Query().Create(category))

	product := models.Product{
		CategoryId: int64(category.ID),
		Name:       "Nasi Goreng",
		IsVariant:  true,
		BasePrice:  15000,
		IsActive:   true,
	}
	s.NoError(facades.Orm().Query().Create(&product))

	variant := models.ProductVariant{
		ProductId: int64(product.ID),
		Name:      "Nasi Goreng",
		Price:     15000,
	}
	s.NoError(facades.Orm().Query().Create(&variant))

	s.NoError(facades.Orm().Query().Create(&models.Stock{
		OutletId:         1,
		ProductVariantId: int64(variant.ID),
		Quantity:         100,
	}))

	return int64(variant.ID)
}

func (s *ReportTestSuite) createTransaction(token, clientUuid string, variantId int64, quantity int) int64 {
	resp, err := s.Http(s.T()).WithToken(token).WithHeader("Content-Type", "application/json").Post("/api/v1/transactions",
		strings.NewReader(fmt.Sprintf(`{
			"client_uuid": "%s",
			"outlet_id": 1,
			"tax_percent": 10,
			"items": [{"product_variant_id": %d, "quantity": %d}],
			"payments": [{"payment_method_id": 1, "amount": 50000}]
		}`, clientUuid, variantId, quantity)),
	)
	s.NoError(err)
	resp.AssertCreated()

	body, err := resp.Json()
	s.NoError(err)

	return int64(body["data"].(map[string]any)["id"].(float64))
}

func (s *ReportTestSuite) daily(token string) map[string]any {
	resp, err := s.Http(s.T()).WithToken(token).Get("/api/v1/reports/daily")
	s.NoError(err)
	resp.AssertOk()

	body, err := resp.Json()
	s.NoError(err)

	return body["data"].(map[string]any)
}

func (s *ReportTestSuite) TestDailyReportAggregatesCompletedOnly() {
	variantId := s.seedProduct()
	kasirToken := s.login("kasir@pos.local")
	adminToken := s.login("admin@pos.local")

	s.createTransaction(kasirToken, "rpt-k1", variantId, 2)
	s.createTransaction(kasirToken, "rpt-k2", variantId, 1)
	adminTrx := s.createTransaction(adminToken, "rpt-a1", variantId, 2)

	adminDaily := s.daily(adminToken)
	s.Equal(int64(3), int64(adminDaily["transaction_count"].(float64)))
	s.Equal(75000.0, adminDaily["subtotal"])
	s.Equal(7500.0, adminDaily["tax"])
	s.Equal(82500.0, adminDaily["total"])
	s.LessOrEqual(1, len(adminDaily["per_day"].([]any)))

	kasirDaily := s.daily(kasirToken)
	s.Equal(int64(2), int64(kasirDaily["transaction_count"].(float64)))
	s.Equal(45000.0, kasirDaily["subtotal"])
	s.Equal(49500.0, kasirDaily["total"])

	resp, err := s.Http(s.T()).WithToken(adminToken).Post(fmt.Sprintf("/api/v1/transactions/%d/void", adminTrx), nil)
	s.NoError(err)
	resp.AssertOk()

	afterVoid := s.daily(adminToken)
	s.Equal(45000.0, afterVoid["subtotal"])
}

func (s *ReportTestSuite) TestProductsReportAggregatesPerProduct() {
	variantId := s.seedProduct()
	kasirToken := s.login("kasir@pos.local")
	adminToken := s.login("admin@pos.local")

	s.createTransaction(kasirToken, "rptp-k1", variantId, 2)
	s.createTransaction(adminToken, "rptp-a1", variantId, 3)

	resp, err := s.Http(s.T()).WithToken(adminToken).Get("/api/v1/reports/products")
	s.NoError(err)
	resp.AssertOk()

	body, err := resp.Json()
	s.NoError(err)
	items := body["data"].(map[string]any)["items"].([]any)
	s.Len(items, 1)

	row := items[0].(map[string]any)
	s.Equal("Nasi Goreng", row["product_name"])
	s.Equal(5.0, row["quantity"])
	s.Equal(75000.0, row["subtotal"])
}

func (s *ReportTestSuite) TestReportRespectsDateRange() {
	variantId := s.seedProduct()
	adminToken := s.login("admin@pos.local")
	s.createTransaction(adminToken, "rptd-1", variantId, 2)

	today := carbon.Now().ToDateString()
	yesterday := carbon.Now().SubDay().ToDateString()

	resp, err := s.Http(s.T()).WithToken(adminToken).Get(fmt.Sprintf("/api/v1/reports/daily?from=%s&to=%s", yesterday, today))
	s.NoError(err)
	resp.AssertOk()
	body, _ := resp.Json()
	s.Equal(1.0, body["data"].(map[string]any)["transaction_count"])
}
