package requests

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type TransactionItemRequest struct {
	ProductVariantId int64   `form:"product_variant_id" json:"product_variant_id"`
	Quantity         float64 `form:"quantity" json:"quantity"`
}

type TransactionPaymentRequest struct {
	PaymentMethodId int64   `form:"payment_method_id" json:"payment_method_id"`
	Amount          float64 `form:"amount" json:"amount"`
}

type TransactionRequest struct {
	ClientUuid string                      `form:"client_uuid" json:"client_uuid"`
	OutletId   int64                       `form:"outlet_id" json:"outlet_id"`
	CustomerId *int64                      `form:"customer_id" json:"customer_id"`
	Discount   float64                     `form:"discount" json:"discount"`
	TaxPercent float64                     `form:"tax_percent" json:"tax_percent"`
	Items      []TransactionItemRequest    `form:"items" json:"items"`
	Payments   []TransactionPaymentRequest `form:"payments" json:"payments"`
}

func (r *TransactionRequest) Authorize(ctx http.Context) error {
	return nil
}

func (r *TransactionRequest) Filters(ctx http.Context) map[string]any {
	return map[string]any{}
}

func (r *TransactionRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"client_uuid": "required",
		"outlet_id":   "required|integer",
		"customer_id": "integer",
		"items":       "required",
		"payments":    "required",
	}
}

func (r *TransactionRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *TransactionRequest) Attributes(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *TransactionRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return nil
}
