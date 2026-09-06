package requests

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type ProductVariantRequest struct {
	Name           string  `form:"name" json:"name"`
	SkuVariant     *string `form:"sku_variant" json:"sku_variant"`
	BarcodeVariant *string `form:"barcode_variant" json:"barcode_variant"`
	Price          float64 `form:"price" json:"price"`
}

type ProductRequest struct {
	CategoryId int64                   `form:"category_id" json:"category_id"`
	Name       string                  `form:"name" json:"name"`
	Sku        *string                 `form:"sku" json:"sku"`
	Barcode    *string                 `form:"barcode" json:"barcode"`
	Unit       *string                 `form:"unit" json:"unit"`
	IsVariant  bool                    `form:"is_variant" json:"is_variant"`
	BasePrice  float64                 `form:"base_price" json:"base_price"`
	IsActive   *bool                   `form:"is_active" json:"is_active"`
	Variants   []ProductVariantRequest `form:"variants" json:"variants"`
}

func (r *ProductRequest) Authorize(ctx http.Context) error {
	return nil
}

func (r *ProductRequest) Filters(ctx http.Context) map[string]any {
	return map[string]any{}
}

func (r *ProductRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"category_id": "required|integer",
		"name":        "required",
		"base_price":  "required",
		"is_variant":  "boolean",
		"is_active":   "boolean",
	}
}

func (r *ProductRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *ProductRequest) Attributes(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *ProductRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return nil
}
