package middleware

import (
	stdhttp "net/http"

	"github.com/goravel/framework/contracts/http"

	"github.com/ridhoauliama97/pos-server/app/facades"
)

type Auth struct{}

func NewAuth() *Auth {
	return &Auth{}
}

func (m *Auth) Signature() string {
	return "Auth"
}

func (m *Auth) Handle(ctx http.Context) {
	token := ctx.Request().Header("Authorization", "")
	if token == "" {
		_ = ctx.Response().Status(stdhttp.StatusUnauthorized).Json(http.Json{
			"error": "unauthorized",
		}).Abort()
		return
	}

	if _, err := facades.Auth(ctx).Parse(token); err != nil {
		_ = ctx.Response().Status(stdhttp.StatusUnauthorized).Json(http.Json{
			"error": "unauthorized",
		}).Abort()
		return
	}

	ctx.Request().Next()
}
