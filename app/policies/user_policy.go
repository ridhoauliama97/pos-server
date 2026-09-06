package policies

import (
	"github.com/goravel/framework/contracts/auth/access"

	"github.com/ridhoauliama97/pos-server/app/models"
)

const (
	AbilityUserView   = "users.view"
	AbilityUserCreate = "users.create"
	AbilityUserUpdate = "users.update"
	AbilityUserDelete = "users.delete"
)

func RegisterUser(gate access.Gate) {
	gate.Define(AbilityUserView, allowRoles(models.RoleOwner, models.RoleAdmin))
	gate.Define(AbilityUserCreate, and(allowRoles(models.RoleOwner, models.RoleAdmin), DenyOwnerRole()))
	gate.Define(AbilityUserUpdate, and(allowRoles(models.RoleOwner, models.RoleAdmin), DenyOwnerRole()))
	gate.Define(AbilityUserDelete, and(allowRoles(models.RoleOwner, models.RoleAdmin), DenyOwnerRole()))
}
