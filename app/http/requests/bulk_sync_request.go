package requests

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type BulkSyncTransactionRequest struct {
	ClientUuid string                      `form:"client_uuid" json:"client_uuid"`
	OutletId   int64                       `form:"outlet_id" json:"outlet_id"`
	CustomerId *int64                      `form:"customer_id" json:"customer_id"`
	Discount   float64                     `form:"discount" json:"discount"`
	TaxPercent float64                     `form:"tax_percent" json:"tax_percent"`
	Items      []TransactionItemRequest    `form:"items" json:"items"`
	Payments   []TransactionPaymentRequest `form:"payments" json:"payments"`
}

type BulkSyncRequest struct {
	Transactions []BulkSyncTransactionRequest `form:"transactions" json:"transactions"`
}

func (r *BulkSyncRequest) Authorize(ctx http.Context) error {
	return nil
}

func (r *BulkSyncRequest) Filters(ctx http.Context) map[string]any {
	return map[string]any{}
}

func (r *BulkSyncRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"transactions": "required|array",
	}
}

func (r *BulkSyncRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *BulkSyncRequest) Attributes(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *BulkSyncRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return nil
}
