package policies

import (
	"github.com/goravel/framework/contracts/auth/access"

	"github.com/ridhoauliama97/pos-server/app/models"
)

const (
	// AbilityReportView scopes to the current user's own shift report for kasir.
	AbilityReportView = "reports.view"
	// AbilityReportViewAll sees reports for every user, outlet or cashier.
	AbilityReportViewAll = "reports.view_all"
)

func RegisterReport(gate access.Gate) {
	gate.Define(AbilityReportView, allowRoles(models.RoleOwner, models.RoleAdmin, models.RoleKasir))
	gate.Define(AbilityReportViewAll, allowRoles(models.RoleOwner, models.RoleAdmin))
}
