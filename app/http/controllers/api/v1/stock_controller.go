package v1

import (
	stdhttp "net/http"

	"github.com/goravel/framework/contracts/http"

	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/http/requests"
	"github.com/ridhoauliama97/pos-server/app/models"
	"github.com/ridhoauliama97/pos-server/app/policies"
	"github.com/ridhoauliama97/pos-server/app/services"
)

type StockController struct{}

func NewStockController() *StockController {
	return &StockController{}
}

func (r *StockController) Index(ctx http.Context) http.Response {
	if !r.authorize(ctx, policies.AbilityStockView) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	query := facades.Orm().Query().Model(&models.Stock{}).
		With("ProductVariant.Product").Order("id asc")
	if outletID := ctx.Request().Query("outlet_id"); outletID != "" {
		query = query.Where("outlet_id = ?", outletID)
	}

	var stocks []models.Stock
	if err := query.Get(&stocks); err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	return ctx.Response().Success().Json(http.Json{"data": r.serializeMany(stocks)})
}

func (r *StockController) Show(ctx http.Context) http.Response {
	if !r.authorize(ctx, policies.AbilityStockView) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	query := facades.Orm().Query().Model(&models.Stock{}).
		With("ProductVariant.Product").
		Where("product_variant_id = ?", ctx.Request().RouteInt64("productVariantId"))
	if outletID := ctx.Request().Query("outlet_id"); outletID != "" {
		query = query.Where("outlet_id = ?", outletID)
	}

	var stock models.Stock
	if err := query.First(&stock); err != nil || stock.ID == 0 {
		return ctx.Response().Status(stdhttp.StatusNotFound).Json(http.Json{"error": "stock not found"})
	}

	return ctx.Response().Success().Json(http.Json{"data": r.serialize(stock)})
}

func (r *StockController) Adjust(ctx http.Context) http.Response {
	var adjustRequest requests.StockAdjustRequest
	invalid, err := ctx.Request().ValidateRequest(&adjustRequest)
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}
	if invalid != nil {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"errors": invalid.All()})
	}

	if !r.authorize(ctx, policies.AbilityStockAdjust) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	if !r.variantExists(adjustRequest.ProductVariantId) {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": "product variant not found"})
	}
	if adjustRequest.Quantity < 0 {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": "quantity must not be negative"})
	}

	var actor models.User
	if err := facades.Auth(ctx).User(&actor); err != nil {
		return ctx.Response().Status(stdhttp.StatusUnauthorized).Json(http.Json{"error": "unauthorized"})
	}
	createdBy := int64(actor.ID)

	stockService := services.NewStockService()
	if err := stockService.AdjustStock(services.StockChange{
		OutletId:         adjustRequest.OutletId,
		ProductVariantId: adjustRequest.ProductVariantId,
		Quantity:         adjustRequest.Quantity,
		Note:             &adjustRequest.Note,
		CreatedBy:        &createdBy,
	}); err != nil {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": err.Error()})
	}

	var stock models.Stock
	if err := facades.Orm().Query().With("ProductVariant.Product").
		Where("outlet_id = ? AND product_variant_id = ?", adjustRequest.OutletId, adjustRequest.ProductVariantId).
		First(&stock); err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	return ctx.Response().Success().Json(http.Json{"data": r.serialize(stock)})
}

func (r *StockController) authorize(ctx http.Context, ability string) bool {
	return facades.Gate().WithContext(ctx).Allows(ability, nil)
}

func (r *StockController) variantExists(id int64) bool {
	count, err := facades.Orm().Query().Model(&models.ProductVariant{}).Where("id = ?", id).Count()
	if err != nil {
		return false
	}

	return count > 0
}

func (r *StockController) serialize(stock models.Stock) http.Json {
	product := http.Json{}
	if stock.ProductVariant != nil && stock.ProductVariant.Product != nil {
		product = http.Json{
			"id":          stock.ProductVariant.Product.ID,
			"name":        stock.ProductVariant.Product.Name,
			"sku":         stock.ProductVariant.Product.Sku,
			"barcode":     stock.ProductVariant.Product.Barcode,
			"category_id": stock.ProductVariant.Product.CategoryId,
		}
	}

	variant := http.Json{"id": stock.ProductVariantId}
	if stock.ProductVariant != nil {
		variant = http.Json{
			"id":    stock.ProductVariant.ID,
			"name":  stock.ProductVariant.Name,
			"price": stock.ProductVariant.Price,
		}
	}

	return http.Json{
		"outlet_id":          stock.OutletId,
		"product_variant_id": stock.ProductVariantId,
		"quantity":           stock.Quantity,
		"min_stock_alert":    stock.MinStockAlert,
		"variant":            variant,
		"product":            product,
	}
}

func (r *StockController) serializeMany(stocks []models.Stock) []http.Json {
	result := make([]http.Json, 0, len(stocks))
	for _, stock := range stocks {
		result = append(result, r.serialize(stock))
	}

	return result
}
