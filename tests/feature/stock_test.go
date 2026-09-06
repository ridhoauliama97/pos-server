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

type StockTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestStockTestSuite(t *testing.T) {
	suite.Run(t, new(StockTestSuite))
}

func (s *StockTestSuite) SetupTest() {
	s.RefreshDatabase()
	s.Seed(&seeders.DatabaseSeeder{})
}

func (s *StockTestSuite) login(email string) string {
	resp, err := s.Http(s.T()).WithHeader("Content-Type", "application/json").Post("/api/v1/auth/login",
		strings.NewReader(`{"email":"`+email+`","password":"password123"}`),
	)
	s.NoError(err)
	resp.AssertOk()
	body, err := resp.Json()
	s.NoError(err)

	return body["token"].(string)
}

func (s *StockTestSuite) seedVariant() int64 {
	category := &models.Category{Name: "Foods"}
	s.NoError(facades.Orm().Query().Create(category))
	product := models.Product{CategoryId: int64(category.ID), Name: "Nasi Goreng", IsVariant: true, BasePrice: 15000, IsActive: true}
	s.NoError(facades.Orm().Query().Create(&product))
	variant := models.ProductVariant{ProductId: int64(product.ID), Name: "Nasi Goreng", Price: 15000}
	s.NoError(facades.Orm().Query().Create(&variant))
	s.NoError(facades.Orm().Query().Create(&models.Stock{OutletId: 1, ProductVariantId: int64(variant.ID), Quantity: 50}))

	return int64(variant.ID)
}

func (s *StockTestSuite) TestConcurrentAdjustSettlesToFinalQuantity() {
	variantId := s.seedVariant()
	token := s.login("admin@pos.local")

	body := fmt.Sprintf(`{"outlet_id":1,"product_variant_id":%d,"quantity":100,"note":"audit"}`, variantId)

	var wg sync.WaitGroup
	successful := make([]bool, 6)
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			resp, err := s.Http(s.T()).WithToken(token).WithHeader("Content-Type", "application/json").Post("/api/v1/stocks/adjust", strings.NewReader(body))
			s.NoError(err)
			successful[index] = resp.IsSuccessful()
		}(i)
	}
	wg.Wait()

	for _, ok := range successful {
		s.True(ok)
	}

	var stock models.Stock
	s.NoError(facades.Orm().Query().Where("outlet_id = ? AND product_variant_id = ?", 1, variantId).First(&stock))
	s.Equal(100.0, stock.Quantity)
}
