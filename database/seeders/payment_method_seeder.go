package seeders

import (
	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/models"
)

type PaymentMethodSeeder struct {
}

// Signature The name and signature of the seeder.
func (s *PaymentMethodSeeder) Signature() string {
	return "PaymentMethodSeeder"
}

// Run executes the seeder logic.
func (s *PaymentMethodSeeder) Run() error {
	if err := facades.Orm().Query().FirstOrCreate(&models.PaymentMethod{
		Code: "cash",
	}, models.PaymentMethod{Code: "cash", Name: "Cash", IsActive: true}); err != nil {
		return err
	}

	return nil
}
