package seeders

import (
	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/models"
)

type OutletSeeder struct {
}

// Signature The name and signature of the seeder.
func (s *OutletSeeder) Signature() string {
	return "OutletSeeder"
}

// Run executes the seeder logic.
func (s *OutletSeeder) Run() error {
	if err := facades.Orm().Query().FirstOrCreate(&models.Outlet{
		Name: "Default Outlet",
	}, models.Outlet{Name: "Default Outlet", IsActive: true}); err != nil {
		return err
	}

	return nil
}
