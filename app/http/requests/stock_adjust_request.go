package requests

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type StockAdjustRequest struct {
	OutletId         int64   `form:"outlet_id" json:"outlet_id"`
	ProductVariantId int64   `form:"product_variant_id" json:"product_variant_id"`
	Quantity         float64 `form:"quantity" json:"quantity"`
	Note             string  `form:"note" json:"note"`
}

func (r *StockAdjustRequest) Authorize(ctx http.Context) error {
	return nil
}

func (r *StockAdjustRequest) Filters(ctx http.Context) map[string]any {
	return map[string]any{}
}

func (r *StockAdjustRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"outlet_id":          "required|integer",
		"product_variant_id": "required|integer",
		"quantity":           "required",
	}
}

func (r *StockAdjustRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *StockAdjustRequest) Attributes(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *StockAdjustRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return nil
}
