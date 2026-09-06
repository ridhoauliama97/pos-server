package services

import (
	"errors"

	frmerrors "github.com/goravel/framework/errors"

	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/models"
)

var ErrCustomerNotFound = errors.New("customer not found")

type CustomerService struct{}

func NewCustomerService() *CustomerService {
	return &CustomerService{}
}

// History returns a paginated list of the customer's completed transactions
// with their item snapshots.
func (s *CustomerService) History(customerId int64, page, limit int) ([]*models.Transaction, int64, error) {
	var customer models.Customer
	if err := facades.Orm().Query().Where("id = ?", customerId).FirstOrFail(&customer); err != nil {
		if errors.Is(err, frmerrors.OrmRecordNotFound) {
			return nil, 0, ErrCustomerNotFound
		}
		return nil, 0, err
	}

	var total int64
	var transactions []*models.Transaction
	query := facades.Orm().Query().
		With("TransactionItems").
		Where("customer_id = ?", customerId).
		Where("status = ?", TransactionStatusCompleted).
		Order("created_at desc")

	if err := query.Paginate(page, limit, &transactions, &total); err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}
