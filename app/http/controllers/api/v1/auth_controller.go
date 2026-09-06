package v1

import (
	"errors"
	stdhttp "net/http"

	"github.com/goravel/framework/contracts/http"
	"gorm.io/gorm"

	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/http/requests"
	"github.com/ridhoauliama97/pos-server/app/models"
)

type AuthController struct{}

func NewAuthController() *AuthController {
	return &AuthController{}
}

func (r *AuthController) Login(ctx http.Context) http.Response {
	var loginRequest requests.LoginRequest
	invalid, err := ctx.Request().ValidateRequest(&loginRequest)
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{
			"error": err.Error(),
		})
	}
	if invalid != nil {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{
			"errors": invalid.All(),
		})
	}

	var user models.User
	if err := facades.Orm().Query().Where("email", loginRequest.Email).FirstOrFail(&user); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Response().Status(stdhttp.StatusUnauthorized).Json(http.Json{
				"error": "invalid credentials",
			})
		}
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{
			"error": err.Error(),
		})
	}

	if !facades.Hash().Check(loginRequest.Password, user.Password) {
		return ctx.Response().Status(stdhttp.StatusUnauthorized).Json(http.Json{
			"error": "invalid credentials",
		})
	}

	if !user.IsActive {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{
			"error": "account is inactive",
		})
	}

	token, err := facades.Auth(ctx).LoginUsingID(user.ID)
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{
			"error": err.Error(),
		})
	}

	return ctx.Response().Status(stdhttp.StatusOK).Json(http.Json{
		"token": token,
		"user": http.Json{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"role":  user.Role,
		},
	})
}

func (r *AuthController) Logout(ctx http.Context) http.Response {
	if err := facades.Auth(ctx).Logout(); err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{
			"error": err.Error(),
		})
	}

	return ctx.Response().Status(stdhttp.StatusOK).Json(http.Json{
		"message": "logged out",
	})
}

func (r *AuthController) Refresh(ctx http.Context) http.Response {
	token, err := facades.Auth(ctx).Refresh()
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusUnauthorized).Json(http.Json{
			"error": err.Error(),
		})
	}

	return ctx.Response().Status(stdhttp.StatusOK).Json(http.Json{
		"token": token,
	})
}
