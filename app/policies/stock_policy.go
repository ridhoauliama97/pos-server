package policies

import (
	"github.com/goravel/framework/contracts/auth/access"

	"github.com/ridhoauliama97/pos-server/app/models"
)

const (
	AbilityStockView   = "stocks.view"
	AbilityStockAdjust = "stocks.adjust"
)

func RegisterStock(gate access.Gate) {
	gate.Define(AbilityStockView, allowRoles(models.RoleOwner, models.RoleAdmin, models.RoleKasir))
	gate.Define(AbilityStockAdjust, allowRoles(models.RoleOwner, models.RoleAdmin))
}
