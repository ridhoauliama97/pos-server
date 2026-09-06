package v1

import (
	stdhttp "net/http"

	"github.com/goravel/framework/contracts/http"

	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/models"
	"github.com/ridhoauliama97/pos-server/app/policies"
	"github.com/ridhoauliama97/pos-server/app/services"
)

type ReceiptController struct{}

func NewReceiptController() *ReceiptController {
	return &ReceiptController{}
}

func (r *ReceiptController) Show(ctx http.Context) http.Response {
	if !r.authorize(ctx, policies.AbilityTransactionView) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	receiptService := services.NewReceiptService()
	transaction, err := receiptService.GetReceipt(ctx.Request().RouteInt64("id"))
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusNotFound).Json(http.Json{"error": err.Error()})
	}

	if !r.canView(ctx, transaction) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	return ctx.Response().Success().Json(http.Json{"data": r.serialize(transaction)})
}

func (r *ReceiptController) authorize(ctx http.Context, ability string) bool {
	return facades.Gate().WithContext(ctx).Allows(ability, nil)
}

func (r *ReceiptController) canView(ctx http.Context, transaction *models.Transaction) bool {
	var actor models.User
	if err := facades.Auth(ctx).User(&actor); err != nil {
		return false
	}

	if actor.Role == models.RoleKasir && transaction.CashierId != int64(actor.ID) {
		return false
	}

	return true
}

func (r *ReceiptController) serialize(transaction *models.Transaction) http.Json {
	outlet := http.Json{}
	if transaction.Outlet != nil {
		outlet = http.Json{
			"id":      transaction.Outlet.ID,
			"name":    transaction.Outlet.Name,
			"address": transaction.Outlet.Address,
			"phone":   transaction.Outlet.Phone,
		}
	}

	cashier := http.Json{}
	if transaction.Cashier != nil {
		cashier = http.Json{
			"id":    transaction.Cashier.ID,
			"name":  transaction.Cashier.Name,
			"email": transaction.Cashier.Email,
		}
	}

	items := make([]http.Json, 0, len(transaction.TransactionItems))
	for _, item := range transaction.TransactionItems {
		items = append(items, http.Json{
			"product_name_snapshot": item.ProductNameSnapshot,
			"product_variant_id":    item.ProductVariantId,
			"price_snapshot":        item.PriceSnapshot,
			"quantity":              item.Quantity,
			"subtotal":              item.Subtotal,
		})
	}

	payments := make([]http.Json, 0, len(transaction.TransactionPayments))
	var tendered float64
	var change float64
	for _, payment := range transaction.TransactionPayments {
		tendered += payment.Amount
		change += payment.ChangeAmount
		paymentMethod := http.Json{"id": payment.PaymentMethodId}
		if payment.PaymentMethod != nil {
			paymentMethod = http.Json{
				"id":   payment.PaymentMethod.ID,
				"code": payment.PaymentMethod.Code,
				"name": payment.PaymentMethod.Name,
			}
		}
		payments = append(payments, http.Json{
			"payment_method": paymentMethod,
			"amount":         payment.Amount,
			"change_amount":  payment.ChangeAmount,
		})
	}

	return http.Json{
		"transaction_id":    transaction.Id,
		"client_uuid":       transaction.ClientUuid,
		"invoice_number":    transaction.InvoiceNumber,
		"status":            transaction.Status,
		"outlet":            outlet,
		"cashier":           cashier,
		"customer_id":       transaction.CustomerId,
		"subtotal":          transaction.Subtotal,
		"discount":          transaction.Discount,
		"tax":               transaction.Tax,
		"total":             transaction.Total,
		"tendered":          tendered,
		"change":            change,
		"client_created_at": transaction.ClientCreatedAt,
		"created_at":        transaction.CreatedAt,
		"items":             items,
		"payments":          payments,
	}
}
