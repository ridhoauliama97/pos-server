package middleware

import (
	stdhttp "net/http"

	"github.com/goravel/framework/contracts/http"

	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/models"
)

type Role struct {
	roles []string
}

func NewRole(roles ...string) *Role {
	return &Role{roles: roles}
}

func (m *Role) Signature() string {
	return "Role"
}

func (m *Role) Handle(ctx http.Context) {
	var user models.User
	if err := facades.Auth(ctx).User(&user); err != nil {
		_ = ctx.Response().Status(stdhttp.StatusUnauthorized).Json(http.Json{
			"error": "unauthorized",
		}).Abort()
		return
	}

	for _, role := range m.roles {
		if user.Role == role {
			ctx.Request().Next()
			return
		}
	}

	_ = ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{
		"error": "forbidden",
	}).Abort()
}
