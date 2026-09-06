package services

import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"strings"
	"time"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	frmerrors "github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/carbon"

	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/models"
)

const (
	TransactionReferenceType   = "transaction"
	TransactionStatusCompleted = "completed"
	TransactionStatusVoided    = "voided"
	TransactionStatusRefunded  = "refunded"
)

var (
	ErrClientUuidRequired     = errors.New("client_uuid is required")
	ErrEmptyItems             = errors.New("transaction must contain at least one item")
	ErrInvalidQuantity        = errors.New("item quantity must be greater than zero")
	ErrProductVariantNotFound = errors.New("product variant not found")
	ErrEmptyPayments          = errors.New("transaction must contain at least one payment")
	ErrInvalidPaymentAmount   = errors.New("payment amount must be greater than zero")
	ErrPaymentBelowTotal      = errors.New("total paid is less than total")
	ErrInvalidDiscount        = errors.New("discount exceeds subtotal")
	ErrPaymentMethodNotFound  = errors.New("payment method not found")
	ErrTransactionIdempotency = errors.New("failed to check existing transaction")
	ErrTransactionNotFound    = errors.New("transaction not found")
	ErrTransactionCannotVoid  = errors.New("only completed transactions can be voided or refunded")
)

func roundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}

type TransactionItemInput struct {
	ProductVariantId int64
	Quantity         float64
}

type TransactionPaymentInput struct {
	PaymentMethodId int64
	Amount          float64
}

type CreateTransactionInput struct {
	OutletId        int64
	CashierId       int64
	CustomerId      *int64
	ClientUuid      string
	Discount        float64
	TaxPercent      float64
	Payments        []TransactionPaymentInput
	Items           []TransactionItemInput
	ClientCreatedAt carbon.DateTime
}

type TransactionService struct {
	stockService *StockService
}

func NewTransactionService() *TransactionService {
	return &TransactionService{
		stockService: NewStockService(),
	}
}

func (s *TransactionService) CreateTransaction(input CreateTransactionInput) (*models.Transaction, bool, error) {
	if input.ClientUuid == "" {
		return nil, false, ErrClientUuidRequired
	}
	if len(input.Items) == 0 {
		return nil, false, ErrEmptyItems
	}
	if len(input.Payments) == 0 {
		return nil, false, ErrEmptyPayments
	}
	for _, payment := range input.Payments {
		if payment.Amount <= 0 {
			return nil, false, ErrInvalidPaymentAmount
		}
	}

	tx, err := facades.Orm().Query().BeginTransaction()
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback()

	var existing models.Transaction
	err = tx.Where("client_uuid = ?", input.ClientUuid).FirstOrFail(&existing)
	if err == nil {
		if err := tx.Commit(); err != nil {
			return nil, false, err
		}
		return s.loadTransactionDetails(&existing), false, nil
	}
	if !errors.Is(err, frmerrors.OrmRecordNotFound) {
		return nil, false, err
	}

	variants, err := s.loadVariants(tx, input.Items)
	if err != nil {
		return nil, false, err
	}

	items := make([]*models.TransactionItem, 0, len(input.Items))
	for _, item := range input.Items {
		if item.Quantity <= 0 {
			return nil, false, ErrInvalidQuantity
		}
		variant, ok := variants[item.ProductVariantId]
		if !ok {
			return nil, false, ErrProductVariantNotFound
		}
		price := roundMoney(variant.Price)
		id := int64(variant.ID)
		items = append(items, &models.TransactionItem{
			ProductVariantId:    &id,
			ProductNameSnapshot: variant.Product.Name,
			PriceSnapshot:       price,
			Quantity:            roundMoney(item.Quantity),
			Subtotal:            roundMoney(price * item.Quantity),
		})
	}

	subtotal := roundMoney(sumSubtotal(items))
	discount := roundMoney(input.Discount)
	if discount > subtotal {
		return nil, false, ErrInvalidDiscount
	}
	tax := roundMoney((subtotal - discount) * input.TaxPercent / 100)
	total := roundMoney(subtotal - discount + tax)

	totalPaid := roundMoney(sumPayments(input.Payments))
	if totalPaid < total {
		return nil, false, ErrPaymentBelowTotal
	}

	var paymentMethods []models.PaymentMethod
	for _, payment := range input.Payments {
		var paymentMethod models.PaymentMethod
		err = tx.Where("id = ? AND is_active = ?", payment.PaymentMethodId, true).FirstOrFail(&paymentMethod)
		if err != nil {
			if errors.Is(err, frmerrors.OrmRecordNotFound) {
				return nil, false, ErrPaymentMethodNotFound
			}
			return nil, false, err
		}
		paymentMethods = append(paymentMethods, paymentMethod)
	}

	clientCreatedAt := input.ClientCreatedAt
	if clientCreatedAt.Carbon == nil {
		clientCreatedAt = *carbon.NewDateTime(carbon.Now())
	}

	var transaction = models.Transaction{
		ClientUuid:      input.ClientUuid,
		OutletId:        input.OutletId,
		CashierId:       input.CashierId,
		CustomerId:      input.CustomerId,
		InvoiceNumber:   ptrString(s.generateInvoiceNumber()),
		Subtotal:        subtotal,
		Discount:        discount,
		Tax:             tax,
		Total:           total,
		Status:          TransactionStatusCompleted,
		ClientCreatedAt: clientCreatedAt,
	}
	if err := tx.Create(&transaction); err != nil {
		// Concurrent requests with the same client_uuid can both pass the
		// existence check and race on the unique index; the loser surfaces a
		// duplicate-key violation here. Replay the winner's transaction instead
		// of failing, so ids stay idempotent.
		if isDuplicateKeyError(err) {
			for i := 0; i < 5; i++ {
				if existing, found := s.findExistingByClientUuid(input.ClientUuid); found {
					return s.loadTransactionDetails(existing), false, nil
				}
				time.Sleep(20 * time.Millisecond)
			}
			return nil, false, ErrTransactionIdempotency
		}
		return nil, false, err
	}

	for _, item := range items {
		item.TransactionId = transaction.Id
	}
	if err := tx.Create(&items); err != nil {
		return nil, false, err
	}

	referenceId := transaction.Id
	referenceType := TransactionReferenceType
	for _, item := range items {
		if err := s.stockService.DecrementStockWithTx(tx, StockChange{
			OutletId:         input.OutletId,
			ProductVariantId: *item.ProductVariantId,
			Quantity:         item.Quantity,
			ReferenceType:    &referenceType,
			ReferenceId:      &referenceId,
			CreatedBy:        &input.CashierId,
		}); err != nil {
			return nil, false, err
		}
	}

	paidAt := *carbon.NewDateTime(carbon.Now())
	changeAmount := roundMoney(totalPaid - total)
	for index, payment := range input.Payments {
		transactionPayment := models.TransactionPayment{
			TransactionId:   transaction.Id,
			PaymentMethodId: payment.PaymentMethodId,
			Amount:          roundMoney(payment.Amount),
			ChangeAmount:    0,
			PaidAt:          paidAt,
		}
		if index == len(input.Payments)-1 {
			transactionPayment.ChangeAmount = changeAmount
		}
		if err := tx.Create(&transactionPayment); err != nil {
			return nil, false, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, false, err
	}

	return s.loadTransactionDetails(&transaction), true, nil
}

func (s *TransactionService) VoidTransaction(transactionId, actorId int64) (*models.Transaction, error) {
	return s.transitionStatus(transactionId, actorId, TransactionStatusVoided)
}

func (s *TransactionService) RefundTransaction(transactionId, actorId int64) (*models.Transaction, error) {
	return s.transitionStatus(transactionId, actorId, TransactionStatusRefunded)
}

func (s *TransactionService) transitionStatus(transactionId, actorId int64, status string) (*models.Transaction, error) {
	tx, err := facades.Orm().Query().BeginTransaction()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var transaction models.Transaction
	err = tx.Where("id = ?", transactionId).LockForUpdate().FirstOrFail(&transaction)
	if err != nil {
		if errors.Is(err, frmerrors.OrmRecordNotFound) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}

	if transaction.Status != TransactionStatusCompleted {
		return nil, ErrTransactionCannotVoid
	}

	if result, err := tx.Model(&models.Transaction{}).
		Where("id = ? AND status = ?", transactionId, TransactionStatusCompleted).
		Update("status", status); err != nil {
		return nil, err
	} else if result.RowsAffected == 0 {
		return nil, ErrTransactionCannotVoid
	}

	var items []*models.TransactionItem
	if err := tx.Where("transaction_id = ?", transactionId).Get(&items); err != nil {
		return nil, err
	}

	referenceType := TransactionReferenceType
	for _, item := range items {
		if item.ProductVariantId == nil {
			continue
		}
		if err := s.stockService.ReturnStockWithTx(tx, StockChange{
			OutletId:         transaction.OutletId,
			ProductVariantId: *item.ProductVariantId,
			Quantity:         item.Quantity,
			ReferenceType:    &referenceType,
			ReferenceId:      &transactionId,
			CreatedBy:        &actorId,
		}); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.loadTransactionDetails(&transaction), nil
}

func (s *TransactionService) loadVariants(tx contractsorm.Query, items []TransactionItemInput) (map[int64]*models.ProductVariant, error) {
	ids := make([]int64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ProductVariantId)
	}

	var variants []*models.ProductVariant
	if err := tx.With("Product").Where("id IN ?", ids).Get(&variants); err != nil {
		return nil, err
	}

	result := make(map[int64]*models.ProductVariant, len(variants))
	for _, variant := range variants {
		result[int64(variant.ID)] = variant
	}

	return result, nil
}

func ptrString(v string) *string {
	return &v
}

func (s *TransactionService) loadTransactionDetails(transaction *models.Transaction) *models.Transaction {
	var result models.Transaction
	if err := facades.Orm().Query().
		With("TransactionItems").
		With("TransactionPayments").
		Where("id = ?", transaction.Id).
		FirstOrFail(&result); err != nil {
		return transaction
	}

	return &result
}

func (s *TransactionService) generateInvoiceNumber() string {
	now := carbon.Now()
	return fmt.Sprintf("INV-%s-%04d", now.Format("20060102150405"), rand.IntN(10000))
}

func sumSubtotal(items []*models.TransactionItem) float64 {
	var total float64
	for _, item := range items {
		total += item.Subtotal
	}

	return total
}

func sumPayments(payments []TransactionPaymentInput) float64 {
	var total float64
	for _, payment := range payments {
		total += payment.Amount
	}

	return total
}

type BulkSyncResult struct {
	ClientUuid  string
	Created     bool
	Skipped     bool
	Failed      bool
	Error       string
	Transaction *models.Transaction
}

// CreateTransactionsBulk applies an ordered list of transactions idempotently.
// Every input that carries a client_uuid that already exists is skipped instead
// of failing, matching the single-transaction replay behaviour.
func (s *TransactionService) CreateTransactionsBulk(inputs []CreateTransactionInput, actorId int64) []BulkSyncResult {
	results := make([]BulkSyncResult, 0, len(inputs))
	for _, input := range inputs {
		if input.CashierId == 0 {
			input.CashierId = actorId
		}

		result := BulkSyncResult{ClientUuid: input.ClientUuid}
		transaction, created, err := s.CreateTransaction(input)
		switch {
		case err != nil:
			result.Failed = true
			result.Error = err.Error()
		case created:
			result.Created = true
			result.Transaction = transaction
		default:
			result.Skipped = true
			result.Transaction = transaction
		}
		results = append(results, result)
	}

	return results
}

func (s *TransactionService) findExistingByClientUuid(clientUuid string) (*models.Transaction, bool) {
	var transaction models.Transaction
	if err := facades.Orm().Query().Where("client_uuid = ?", clientUuid).FirstOrFail(&transaction); err != nil {
		return nil, false
	}

	return &transaction, true
}

// isDuplicateKeyError reports a unique-constraint violation surfaced by the
// PostgreSQL driver (SQLSTATE 23505) or its gorm translation.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())

	return strings.Contains(message, "sqlstate 23505") ||
		strings.Contains(message, "duplicate key value violates unique constraint") ||
		strings.Contains(message, "duplicated key not allowed")
}
