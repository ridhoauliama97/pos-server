package policies

import (
	"github.com/goravel/framework/contracts/auth/access"

	"github.com/ridhoauliama97/pos-server/app/models"
)

const (
	AbilityCustomerView   = "customers.view"
	AbilityCustomerCreate = "customers.create"
	AbilityCustomerUpdate = "customers.update"
	AbilityCustomerDelete = "customers.delete"
)

func RegisterCustomer(gate access.Gate) {
	gate.Define(AbilityCustomerView, allowRoles(models.RoleOwner, models.RoleAdmin, models.RoleKasir))
	gate.Define(AbilityCustomerCreate, allowRoles(models.RoleOwner, models.RoleAdmin, models.RoleKasir))
	gate.Define(AbilityCustomerUpdate, allowRoles(models.RoleOwner, models.RoleAdmin))
	gate.Define(AbilityCustomerDelete, allowRoles(models.RoleOwner, models.RoleAdmin))
}
