package policies

import (
	"github.com/goravel/framework/contracts/auth/access"

	"github.com/ridhoauliama97/pos-server/app/models"
)

const (
	AbilityTransactionView   = "transactions.view"
	AbilityTransactionCreate = "transactions.create"
	AbilityTransactionVoid   = "transactions.void"
)

func RegisterTransaction(gate access.Gate) {
	gate.Define(AbilityTransactionView, allowRoles(models.RoleOwner, models.RoleAdmin, models.RoleKasir))
	gate.Define(AbilityTransactionCreate, allowRoles(models.RoleOwner, models.RoleAdmin, models.RoleKasir))
	gate.Define(AbilityTransactionVoid, allowRoles(models.RoleOwner, models.RoleAdmin))
}
