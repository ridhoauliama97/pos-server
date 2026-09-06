package services

import (
	"errors"

	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/models"
)

var ErrReceiptNotFound = errors.New("receipt not found")

type ReceiptService struct{}

func NewReceiptService() *ReceiptService {
	return &ReceiptService{}
}

func (s *ReceiptService) GetReceipt(transactionId int64) (*models.Transaction, error) {
	var transaction models.Transaction
	if err := facades.Orm().Query().
		With("Outlet").
		With("Cashier").
		With("TransactionItems").
		With("TransactionPayments.PaymentMethod").
		Where("id = ?", transactionId).
		First(&transaction); err != nil || transaction.Id == 0 {
		return nil, ErrReceiptNotFound
	}

	return &transaction, nil
}
