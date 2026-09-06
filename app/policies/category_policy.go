package policies

import (
	"github.com/goravel/framework/contracts/auth/access"

	"github.com/ridhoauliama97/pos-server/app/models"
)

const (
	AbilityCategoryView   = "categories.view"
	AbilityCategoryCreate = "categories.create"
	AbilityCategoryUpdate = "categories.update"
	AbilityCategoryDelete = "categories.delete"
)

func RegisterCategory(gate access.Gate) {
	gate.Define(AbilityCategoryView, allowRoles(models.RoleOwner, models.RoleAdmin, models.RoleKasir))
	gate.Define(AbilityCategoryCreate, allowRoles(models.RoleOwner, models.RoleAdmin))
	gate.Define(AbilityCategoryUpdate, allowRoles(models.RoleOwner, models.RoleAdmin))
	gate.Define(AbilityCategoryDelete, allowRoles(models.RoleOwner, models.RoleAdmin))
}
