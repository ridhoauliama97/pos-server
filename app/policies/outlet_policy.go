package policies

import (
	"github.com/goravel/framework/contracts/auth/access"

	"github.com/ridhoauliama97/pos-server/app/models"
)

const (
	AbilityOutletView   = "outlets.view"
	AbilityOutletUpdate = "outlets.update"
)

func RegisterOutlet(gate access.Gate) {
	gate.Define(AbilityOutletView, allowRoles(models.RoleOwner))
	gate.Define(AbilityOutletUpdate, allowRoles(models.RoleOwner))
}
