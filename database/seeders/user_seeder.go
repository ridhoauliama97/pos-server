package seeders

import (
	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/models"
)

type UserSeeder struct {
}

// Signature The name and signature of the seeder.
func (s *UserSeeder) Signature() string {
	return "UserSeeder"
}

// Run executes the seeder logic.
func (s *UserSeeder) Run() error {
	count, err := facades.Orm().Query().Model(&models.User{}).Where("email", "admin@pos.local").Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	password, err := facades.Hash().Make("password123")
	if err != nil {
		return err
	}

	return facades.Orm().Query().Create(&models.User{
		Email:    "admin@pos.local",
		Name:     "Admin",
		Password: password,
		Role:     "admin",
		IsActive: true,
	})
}
