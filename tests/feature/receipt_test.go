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

type ReceiptTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestReceiptTestSuite(t *testing.T) {
	suite.Run(t, new(ReceiptTestSuite))
}

func (s *ReceiptTestSuite) SetupTest() {
	s.RefreshDatabase()
	s.Seed(&seeders.DatabaseSeeder{})
}

func (s *ReceiptTestSuite) seedKasir(name, email string) {
	password, err := facades.Hash().Make("password123")
	s.NoError(err)
	s.NoError(facades.Orm().Query().Create(&models.User{
		Name:     name,
		Email:    email,
		Password: password,
		Role:     models.RoleKasir,
		IsActive: true,
	}))
}

func (s *ReceiptTestSuite) login(email string) string {
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

func (s *ReceiptTestSuite) seedProduct() int64 {
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
		Quantity:         50,
	}))

	return int64(variant.ID)
}

func (s *ReceiptTestSuite) createTransaction(token, clientUuid string, variantId int64) int64 {
	resp, err := s.Http(s.T()).WithToken(token).WithHeader("Content-Type", "application/json").Post("/api/v1/transactions",
		strings.NewReader(fmt.Sprintf(`{
			"client_uuid": "%s",
			"outlet_id": 1,
			"tax_percent": 10,
			"items": [
				{"product_variant_id": %d, "quantity": 2},
				{"product_variant_id": %d, "quantity": 1}
			],
			"payments": [{"payment_method_id": 1, "amount": 50000}]
		}`, clientUuid, variantId, variantId)),
	)
	s.NoError(err)
	resp.AssertCreated()

	body, err := resp.Json()
	s.NoError(err)

	return int64(body["data"].(map[string]any)["id"].(float64))
}

func (s *ReceiptTestSuite) TestReceiptEndpointReturnsStructuredJSON() {
	variantId := s.seedProduct()
	token := s.login("admin@pos.local")
	id := s.createTransaction(token, "rec-1", variantId)

	resp, err := s.Http(s.T()).WithToken(token).Get(fmt.Sprintf("/api/v1/transactions/%d/receipt", id))
	s.NoError(err)
	resp.AssertOk()

	body, err := resp.Json()
	s.NoError(err)
	receipt := body["data"].(map[string]any)

	s.Equal(int64(id), int64(receipt["transaction_id"].(float64)))
	s.Equal("rec-1", receipt["client_uuid"])
	s.True(strings.HasPrefix(receipt["invoice_number"].(string), "INV-"))
	s.Equal("completed", receipt["status"])
	s.Equal(45000.0, receipt["subtotal"])
	s.Equal(0.0, receipt["discount"])
	s.Equal(4500.0, receipt["tax"])
	s.Equal(49500.0, receipt["total"])
	s.Equal(50000.0, receipt["tendered"])
	s.Equal(500.0, receipt["change"])
	s.NotEmpty(receipt["client_created_at"])
	s.NotEmpty(receipt["created_at"])

	outlet := receipt["outlet"].(map[string]any)
	s.Equal(float64(1), outlet["id"])
	s.Equal("Default Outlet", outlet["name"])

	cashier := receipt["cashier"].(map[string]any)
	s.Equal("Admin", cashier["name"])
	s.Equal("admin@pos.local", cashier["email"])

	items := receipt["items"].([]any)
	s.Len(items, 2)
	item := items[0].(map[string]any)
	s.Equal("Nasi Goreng", item["product_name_snapshot"])
	s.Equal(15000.0, item["price_snapshot"])
	s.Equal(2.0, item["quantity"])
	s.Equal(30000.0, item["subtotal"])

	payments := receipt["payments"].([]any)
	s.Len(payments, 1)
	payment := payments[0].(map[string]any)
	paymentMethod := payment["payment_method"].(map[string]any)
	s.Equal("cash", paymentMethod["code"])
	s.Equal("Cash", paymentMethod["name"])
	s.Equal(50000.0, payment["amount"])
	s.Equal(500.0, payment["change_amount"])
}

func (s *ReceiptTestSuite) TestKasirSeesOnlyOwnReceipt() {
	s.seedKasir("Kasir A", "kasir@pos.local")
	s.seedKasir("Kasir B", "kasir-b@pos.local")
	variantId := s.seedProduct()
	kasirAToken := s.login("kasir@pos.local")
	kasirBToken := s.login("kasir-b@pos.local")
	adminToken := s.login("admin@pos.local")

	adminId := s.createTransaction(adminToken, "rec-owner-1", variantId)
	kasirBId := s.createTransaction(kasirBToken, "rec-b-1", variantId)

	resp, err := s.Http(s.T()).WithToken(kasirAToken).Get(fmt.Sprintf("/api/v1/transactions/%d/receipt", adminId))
	s.NoError(err)
	resp.AssertForbidden()

	resp, err = s.Http(s.T()).WithToken(kasirAToken).Get(fmt.Sprintf("/api/v1/transactions/%d/receipt", kasirBId))
	s.NoError(err)
	resp.AssertForbidden()

	resp, err = s.Http(s.T()).WithToken(kasirBToken).Get(fmt.Sprintf("/api/v1/transactions/%d/receipt", kasirBId))
	s.NoError(err)
	resp.AssertOk()

	resp, err = s.Http(s.T()).WithToken(adminToken).Get(fmt.Sprintf("/api/v1/transactions/%d/receipt", kasirBId))
	s.NoError(err)
	resp.AssertOk()
}

func (s *ReceiptTestSuite) TestReceiptNotFound() {
	token := s.login("admin@pos.local")

	resp, err := s.Http(s.T()).WithToken(token).Get("/api/v1/transactions/999999/receipt")
	s.NoError(err)
	resp.AssertStatus(404)
}
