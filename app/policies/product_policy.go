package policies

import (
	"github.com/goravel/framework/contracts/auth/access"

	"github.com/ridhoauliama97/pos-server/app/models"
)

const (
	AbilityProductView   = "products.view"
	AbilityProductCreate = "products.create"
	AbilityProductUpdate = "products.update"
	AbilityProductDelete = "products.delete"
)

func RegisterProduct(gate access.Gate) {
	gate.Define(AbilityProductView, allowRoles(models.RoleOwner, models.RoleAdmin, models.RoleKasir))
	gate.Define(AbilityProductCreate, allowRoles(models.RoleOwner, models.RoleAdmin))
	gate.Define(AbilityProductUpdate, allowRoles(models.RoleOwner, models.RoleAdmin))
	gate.Define(AbilityProductDelete, allowRoles(models.RoleOwner, models.RoleAdmin))
}
