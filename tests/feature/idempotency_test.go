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

type IdempotencyTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestIdempotencyTestSuite(t *testing.T) {
	suite.Run(t, new(IdempotencyTestSuite))
}

func (s *IdempotencyTestSuite) SetupTest() {
	s.RefreshDatabase()
	s.Seed(&seeders.DatabaseSeeder{})
}

func (s *IdempotencyTestSuite) login(email string) string {
	resp, err := s.Http(s.T()).WithHeader("Content-Type", "application/json").Post("/api/v1/auth/login",
		strings.NewReader(`{"email":"`+email+`","password":"password123"}`),
	)
	s.NoError(err)
	resp.AssertOk()
	body, _ := resp.Json()

	return body["token"].(string)
}

func (s *IdempotencyTestSuite) seedVariant() int64 {
	category := &models.Category{Name: "Foods"}
	s.NoError(facades.Orm().Query().Create(category))
	product := models.Product{CategoryId: int64(category.ID), Name: "Nasi Goreng", IsVariant: true, BasePrice: 15000, IsActive: true}
	s.NoError(facades.Orm().Query().Create(&product))
	variant := models.ProductVariant{ProductId: int64(product.ID), Name: "Nasi Goreng", Price: 15000}
	s.NoError(facades.Orm().Query().Create(&variant))
	s.NoError(facades.Orm().Query().Create(&models.Stock{OutletId: 1, ProductVariantId: int64(variant.ID), Quantity: 50}))

	return int64(variant.ID)
}

func (s *IdempotencyTestSuite) countTransactions(uuid string) int64 {
	count, err := facades.Orm().Query().Model(&models.Transaction{}).Where("client_uuid = ?", uuid).Count()
	s.NoError(err)

	return count
}

func (s *IdempotencyTestSuite) TestConcurrentSameUuidDoesNotDoubleInsert() {
	variantId := s.seedVariant()
	token := s.login("admin@pos.local")

	body := fmt.Sprintf(`{
		"client_uuid": "idem-race-1", "outlet_id": 1, "tax_percent": 10,
		"items": [{"product_variant_id": %d, "quantity": 2}],
		"payments": [{"payment_method_id": 1, "amount": 50000}]
	}`, variantId)

	var wg sync.WaitGroup
	successful := make([]bool, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			resp, err := s.Http(s.T()).WithToken(token).WithHeader("Content-Type", "application/json").Post("/api/v1/transactions", strings.NewReader(body))
			s.NoError(err)
			successful[index] = resp.IsSuccessful()
		}(i)
	}
	wg.Wait()

	for _, ok := range successful {
		s.True(ok)
	}

	s.Equal(int64(1), s.countTransactions("idem-race-1"))

	var stock models.Stock
	s.NoError(facades.Orm().Query().Where("outlet_id = ? AND product_variant_id = ?", 1, variantId).First(&stock))
	s.Equal(48.0, stock.Quantity)
}

func (s *IdempotencyTestSuite) TestBulkSyncCreatesSkipsAndReportsFailures() {
	variantId := s.seedVariant()
	token := s.login("admin@pos.local")

	// prime one transaction that will be re-sent
	resp, err := s.Http(s.T()).WithToken(token).WithHeader("Content-Type", "application/json").Post("/api/v1/transactions",
		strings.NewReader(fmt.Sprintf(`{
			"client_uuid": "bulk-existing", "outlet_id": 1, "tax_percent": 10,
			"items": [{"product_variant_id": %d, "quantity": 1}],
			"payments": [{"payment_method_id": 1, "amount": 50000}]
		}`, variantId)),
	)
	s.NoError(err)
	resp.AssertCreated()

	payload := fmt.Sprintf(`{
		"transactions": [
			{
				"client_uuid": "bulk-existing", "outlet_id": 1, "tax_percent": 10,
				"items": [{"product_variant_id": %d, "quantity": 1}],
				"payments": [{"payment_method_id": 1, "amount": 50000}]
			},
			{
				"client_uuid": "bulk-new", "outlet_id": 1, "tax_percent": 10,
				"items": [{"product_variant_id": %d, "quantity": 1}],
				"payments": [{"payment_method_id": 1, "amount": 50000}]
			},
			{
				"client_uuid": "bulk-bad", "outlet_id": 1, "tax_percent": 10,
				"items": [{"product_variant_id": %d, "quantity": 1}],
				"payments": [{"payment_method_id": 1, "amount": 1}]
			}
		]
	}`, variantId, variantId, variantId)

	resp, err = s.Http(s.T()).WithToken(token).WithHeader("Content-Type", "application/json").Post("/api/v1/transactions/bulk-sync", strings.NewReader(payload))
	s.NoError(err)
	resp.AssertOk()

	body, err := resp.Json()
	s.NoError(err)
	data := body["data"].(map[string]any)
	s.Equal(1.0, data["created"])
	s.Equal(1.0, data["skipped"])
	s.Equal(1.0, data["failed"])

	results := data["results"].([]any)
	byUuid := map[string]map[string]any{}
	for _, item := range results {
		row := item.(map[string]any)
		byUuid[row["client_uuid"].(string)] = row
	}
	s.Equal("skipped", byUuid["bulk-existing"]["status"])
	s.Equal("created", byUuid["bulk-new"]["status"])
	s.Equal("failed", byUuid["bulk-bad"]["status"])
	s.NotEmpty(byUuid["bulk-bad"]["error"])

	s.Equal(int64(1), s.countTransactions("bulk-new"))
}
