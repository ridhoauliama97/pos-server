package v1

import (
	"errors"
	stdhttp "net/http"

	"github.com/goravel/framework/contracts/http"

	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/http/requests"
	"github.com/ridhoauliama97/pos-server/app/models"
	"github.com/ridhoauliama97/pos-server/app/policies"
	"github.com/ridhoauliama97/pos-server/app/services"
)

type TransactionController struct{}

func NewTransactionController() *TransactionController {
	return &TransactionController{}
}

func (r *TransactionController) Store(ctx http.Context) http.Response {
	var request requests.TransactionRequest
	invalid, err := ctx.Request().ValidateRequest(&request)
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}
	if invalid != nil {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"errors": invalid.All()})
	}

	if !r.authorize(ctx, policies.AbilityTransactionCreate) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	if len(request.Items) == 0 {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": "items is required"})
	}
	if len(request.Payments) == 0 {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": "payments is required"})
	}
	for _, item := range request.Items {
		if item.Quantity <= 0 {
			return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": "item quantity must be greater than zero"})
		}
	}

	var actor models.User
	if err := facades.Auth(ctx).User(&actor); err != nil {
		return ctx.Response().Status(stdhttp.StatusUnauthorized).Json(http.Json{"error": "unauthorized"})
	}

	items := make([]services.TransactionItemInput, len(request.Items))
	for index, item := range request.Items {
		items[index] = services.TransactionItemInput{
			ProductVariantId: item.ProductVariantId,
			Quantity:         item.Quantity,
		}
	}

	payments := make([]services.TransactionPaymentInput, len(request.Payments))
	for index, payment := range request.Payments {
		payments[index] = services.TransactionPaymentInput{
			PaymentMethodId: payment.PaymentMethodId,
			Amount:          payment.Amount,
		}
	}

	service := services.NewTransactionService()
	transaction, created, err := service.CreateTransaction(services.CreateTransactionInput{
		OutletId:   request.OutletId,
		CashierId:  int64(actor.ID),
		CustomerId: request.CustomerId,
		ClientUuid: request.ClientUuid,
		Discount:   request.Discount,
		TaxPercent: request.TaxPercent,
		Items:      items,
		Payments:   payments,
	})
	if err != nil {
		return ctx.Response().Status(r.statusForError(err)).Json(http.Json{"error": err.Error()})
	}

	status := stdhttp.StatusCreated
	if !created {
		status = stdhttp.StatusOK
	}

	return ctx.Response().Status(status).Json(http.Json{"data": r.serialize(transaction)})
}

func (r *TransactionController) Void(ctx http.Context) http.Response {
	return r.transition(ctx, true)
}

func (r *TransactionController) Refund(ctx http.Context) http.Response {
	return r.transition(ctx, false)
}

func (r *TransactionController) BulkSync(ctx http.Context) http.Response {
	var request requests.BulkSyncRequest
	invalid, err := ctx.Request().ValidateRequest(&request)
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}
	if invalid != nil {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"errors": invalid.All()})
	}

	if !r.authorize(ctx, policies.AbilityTransactionCreate) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	var actor models.User
	if err := facades.Auth(ctx).User(&actor); err != nil {
		return ctx.Response().Status(stdhttp.StatusUnauthorized).Json(http.Json{"error": "unauthorized"})
	}

	inputs := make([]services.CreateTransactionInput, 0, len(request.Transactions))
	for _, transactionRequest := range request.Transactions {
		items := make([]services.TransactionItemInput, len(transactionRequest.Items))
		for index, item := range transactionRequest.Items {
			items[index] = services.TransactionItemInput{
				ProductVariantId: item.ProductVariantId,
				Quantity:         item.Quantity,
			}
		}

		payments := make([]services.TransactionPaymentInput, len(transactionRequest.Payments))
		for index, payment := range transactionRequest.Payments {
			payments[index] = services.TransactionPaymentInput{
				PaymentMethodId: payment.PaymentMethodId,
				Amount:          payment.Amount,
			}
		}

		inputs = append(inputs, services.CreateTransactionInput{
			OutletId:   transactionRequest.OutletId,
			CustomerId: transactionRequest.CustomerId,
			ClientUuid: transactionRequest.ClientUuid,
			Discount:   transactionRequest.Discount,
			TaxPercent: transactionRequest.TaxPercent,
			Items:      items,
			Payments:   payments,
		})
	}

	service := services.NewTransactionService()
	results := service.CreateTransactionsBulk(inputs, int64(actor.ID))

	items := make([]http.Json, 0, len(results))
	var created, skipped, failed int
	for _, result := range results {
		status := "skipped"
		switch {
		case result.Failed:
			status = "failed"
			failed++
		case result.Created:
			status = "created"
			created++
		default:
			skipped++
		}

		item := http.Json{
			"client_uuid": result.ClientUuid,
			"status":      status,
		}
		if result.Failed {
			item["error"] = result.Error
		} else {
			item["transaction_id"] = result.Transaction.Id
		}
		items = append(items, item)
	}

	return ctx.Response().Success().Json(http.Json{
		"data": http.Json{
			"created": created,
			"skipped": skipped,
			"failed":  failed,
			"results": items,
		},
	})
}

func (r *TransactionController) transition(ctx http.Context, void bool) http.Response {
	if !r.authorize(ctx, policies.AbilityTransactionVoid) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	var actor models.User
	if err := facades.Auth(ctx).User(&actor); err != nil {
		return ctx.Response().Status(stdhttp.StatusUnauthorized).Json(http.Json{"error": "unauthorized"})
	}

	service := services.NewTransactionService()
	action := service.RefundTransaction
	if void {
		action = service.VoidTransaction
	}

	transaction, err := action(ctx.Request().RouteInt64("id"), int64(actor.ID))
	if err != nil {
		if errors.Is(err, services.ErrTransactionNotFound) {
			return ctx.Response().Status(stdhttp.StatusNotFound).Json(http.Json{"error": err.Error()})
		}
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": err.Error()})
	}

	return ctx.Response().Success().Json(http.Json{"data": r.serialize(transaction)})
}

func (r *TransactionController) authorize(ctx http.Context, ability string) bool {
	return facades.Gate().WithContext(ctx).Allows(ability, nil)
}

func (r *TransactionController) statusForError(err error) int {
	switch {
	case errors.Is(err, services.ErrInsufficientStock),
		errors.Is(err, services.ErrProductVariantNotFound),
		errors.Is(err, services.ErrPaymentMethodNotFound),
		errors.Is(err, services.ErrPaymentBelowTotal),
		errors.Is(err, services.ErrInvalidDiscount),
		errors.Is(err, services.ErrInvalidQuantity),
		errors.Is(err, services.ErrInvalidPaymentAmount),
		errors.Is(err, services.ErrEmptyItems),
		errors.Is(err, services.ErrEmptyPayments),
		errors.Is(err, services.ErrClientUuidRequired):
		return stdhttp.StatusUnprocessableEntity
	default:
		return stdhttp.StatusInternalServerError
	}
}

func (r *TransactionController) serialize(transaction *models.Transaction) http.Json {
	items := make([]http.Json, 0, len(transaction.TransactionItems))
	for _, item := range transaction.TransactionItems {
		items = append(items, http.Json{
			"product_variant_id":    item.ProductVariantId,
			"product_name_snapshot": item.ProductNameSnapshot,
			"price_snapshot":        item.PriceSnapshot,
			"quantity":              item.Quantity,
			"subtotal":              item.Subtotal,
		})
	}

	payments := make([]http.Json, 0, len(transaction.TransactionPayments))
	for _, payment := range transaction.TransactionPayments {
		payments = append(payments, http.Json{
			"payment_method_id": payment.PaymentMethodId,
			"amount":            payment.Amount,
			"change_amount":     payment.ChangeAmount,
			"paid_at":           payment.PaidAt,
		})
	}

	return http.Json{
		"id":                transaction.Id,
		"client_uuid":       transaction.ClientUuid,
		"outlet_id":         transaction.OutletId,
		"cashier_id":        transaction.CashierId,
		"customer_id":       transaction.CustomerId,
		"invoice_number":    transaction.InvoiceNumber,
		"subtotal":          transaction.Subtotal,
		"discount":          transaction.Discount,
		"tax":               transaction.Tax,
		"total":             transaction.Total,
		"status":            transaction.Status,
		"client_created_at": transaction.ClientCreatedAt,
		"created_at":        transaction.CreatedAt,
		"items":             items,
		"payments":          payments,
	}
}
