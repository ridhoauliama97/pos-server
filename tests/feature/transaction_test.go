package feature

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/models"
	"github.com/ridhoauliama97/pos-server/database/seeders"
	"github.com/ridhoauliama97/pos-server/tests"
)

type TransactionTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestTransactionTestSuite(t *testing.T) {
	suite.Run(t, new(TransactionTestSuite))
}

func (s *TransactionTestSuite) SetupTest() {
	s.RefreshDatabase()
	s.Seed(&seeders.DatabaseSeeder{})
}

func (s *TransactionTestSuite) seedKasir(name, email string) {
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

func (s *TransactionTestSuite) login(email string) string {
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

func (s *TransactionTestSuite) seedProduct(price float64, stock int) int64 {
	category := &models.Category{Name: "Foods"}
	s.NoError(facades.Orm().Query().Create(category))

	product := models.Product{
		CategoryId: int64(category.ID),
		Name:       "Nasi Goreng",
		IsVariant:  true,
		BasePrice:  price,
		IsActive:   true,
	}
	s.NoError(facades.Orm().Query().Create(&product))

	variant := models.ProductVariant{
		ProductId: int64(product.ID),
		Name:      "Nasi Goreng",
		Price:     price,
	}
	s.NoError(facades.Orm().Query().Create(&variant))

	s.NoError(facades.Orm().Query().Create(&models.Stock{
		OutletId:         1,
		ProductVariantId: int64(variant.ID),
		Quantity:         float64(stock),
	}))

	return int64(variant.ID)
}

// createTransaction returns the parsed response, transaction id and invoice number.
func (s *TransactionTestSuite) createTransaction(token, clientUuid string, variantId int64, quantity int) (map[string]any, int64, string) {
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
	data := body["data"].(map[string]any)

	return data, int64(data["id"].(float64)), data["invoice_number"].(string)
}

func (s *TransactionTestSuite) stockQuantity(variantId int64) float64 {
	var stock models.Stock
	s.NoError(facades.Orm().Query().Where("outlet_id = ? AND product_variant_id = ?", 1, variantId).First(&stock))

	return stock.Quantity
}

func (s *TransactionTestSuite) movementsFor(referenceId int64) []*models.StockMovement {
	var movements []*models.StockMovement
	s.NoError(facades.Orm().Query().
		Where("reference_type = ? AND reference_id = ?", "transaction", referenceId).
		Order("id asc").
		Get(&movements))

	return movements
}

func (s *TransactionTestSuite) TestCreateTransactionDecrementsStockAndRecordsMovement() {
	variantId := s.seedProduct(15000, 10)
	token := s.login("admin@pos.local")

	data, id, invoice := s.createTransaction(token, "trx-create-1", variantId, 2)

	s.Equal(30000.0, data["subtotal"])
	s.Equal(0.0, data["discount"])
	s.Equal(3000.0, data["tax"])
	s.Equal(33000.0, data["total"])
	s.Equal("completed", data["status"])
	s.True(strings.HasPrefix(invoice, "INV-"))

	s.Equal(8.0, s.stockQuantity(variantId))

	movements := s.movementsFor(id)
	s.Len(movements, 1)
	s.Equal("sale", movements[0].Type)
	s.Equal(2.0, movements[0].Quantity)
	s.Equal("transaction", *movements[0].ReferenceType)
	s.Equal(int64(id), *movements[0].ReferenceId)
	s.NotNil(movements[0].CreatedBy)
}

func (s *TransactionTestSuite) TestDuplicateClientUuidReturnsExistingWithoutDoubleInsert() {
	variantId := s.seedProduct(15000, 10)
	token := s.login("admin@pos.local")

	body, id, _ := s.createTransaction(token, "trx-idem-1", variantId, 2)

	resp, err := s.Http(s.T()).WithToken(token).WithHeader("Content-Type", "application/json").Post("/api/v1/transactions",
		strings.NewReader(`{
			"client_uuid": "trx-idem-1",
			"outlet_id": 1,
			"tax_percent": 10,
			"items": [{"product_variant_id": `+fmt.Sprintf("%d", variantId)+`, "quantity": 2}],
			"payments": [{"payment_method_id": 1, "amount": 50000}]
		}`),
	)
	s.NoError(err)
	resp.AssertOk()

	replay, err := resp.Json()
	s.NoError(err)
	replayData := replay["data"].(map[string]any)
	s.Equal(int64(id), int64(replayData["id"].(float64)))
	s.Equal(body["total"], replayData["total"])

	s.Equal(8.0, s.stockQuantity(variantId))
	s.Len(s.movementsFor(id), 1)
}

func (s *TransactionTestSuite) TestKasirCanCreateButCannotVoidOrRefund() {
	s.seedKasir("Kasir", "kasir@pos.local")
	variantId := s.seedProduct(15000, 10)
	kasirToken := s.login("kasir@pos.local")

	_, id, _ := s.createTransaction(kasirToken, "trx-kasir-1", variantId, 1)

	invalid, err := s.Http(s.T()).WithToken(kasirToken).Post(fmt.Sprintf("/api/v1/transactions/%d/void", id), nil)
	s.NoError(err)
	invalid.AssertForbidden()

	invalid, err = s.Http(s.T()).WithToken(kasirToken).Post(fmt.Sprintf("/api/v1/transactions/%d/refund", id), nil)
	s.NoError(err)
	invalid.AssertForbidden()

	s.Equal(9.0, s.stockQuantity(variantId))
}

func (s *TransactionTestSuite) TestVoidRestoresStockAndUpdatesStatus() {
	variantId := s.seedProduct(15000, 10)
	token := s.login("admin@pos.local")

	_, id, _ := s.createTransaction(token, "trx-void-1", variantId, 2)
	s.Equal(8.0, s.stockQuantity(variantId))

	resp, err := s.Http(s.T()).WithToken(token).Post(fmt.Sprintf("/api/v1/transactions/%d/void", id), nil)
	s.NoError(err)
	resp.AssertOk()

	body, err := resp.Json()
	s.NoError(err)
	s.Equal("voided", body["data"].(map[string]any)["status"])

	s.Equal(10.0, s.stockQuantity(variantId))

	movements := s.movementsFor(id)
	s.Len(movements, 2)
	s.Equal("sale", movements[0].Type)
	s.Equal("return", movements[1].Type)
	s.Equal(2.0, movements[1].Quantity)

	resp, err = s.Http(s.T()).WithToken(token).Post(fmt.Sprintf("/api/v1/transactions/%d/void", id), nil)
	s.NoError(err)
	resp.AssertStatus(422)
}

func (s *TransactionTestSuite) TestRefundUpdatesStatusAndRestoresStock() {
	variantId := s.seedProduct(15000, 10)
	token := s.login("admin@pos.local")

	_, id, _ := s.createTransaction(token, "trx-refund-1", variantId, 1)
	s.Equal(9.0, s.stockQuantity(variantId))

	resp, err := s.Http(s.T()).WithToken(token).Post(fmt.Sprintf("/api/v1/transactions/%d/refund", id), nil)
	s.NoError(err)
	resp.AssertOk()

	body, err := resp.Json()
	s.NoError(err)
	s.Equal("refunded", body["data"].(map[string]any)["status"])

	s.Equal(10.0, s.stockQuantity(variantId))
}

func (s *TransactionTestSuite) TestConcurrentVoidRestoresStockOnce() {
	variantId := s.seedProduct(15000, 10)
	token := s.login("admin@pos.local")

	_, id, _ := s.createTransaction(token, "trx-void-race", variantId, 2)
	s.Equal(8.0, s.stockQuantity(variantId))

	var wg sync.WaitGroup
	successful := make([]bool, 6)
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			resp, err := s.Http(s.T()).WithToken(token).Post(fmt.Sprintf("/api/v1/transactions/%d/void", id), nil)
			s.NoError(err)
			successful[index] = resp.IsSuccessful()
		}(i)
	}
	wg.Wait()

	successCount := 0
	for _, ok := range successful {
		if ok {
			successCount++
		}
	}
	s.Equal(1, successCount)
	s.Equal(10.0, s.stockQuantity(variantId))
	s.Len(s.movementsFor(id), 2)
}
