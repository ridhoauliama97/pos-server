package requests

import (
	stdhttp "net/http"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type UserRequest struct {
	Name     string `form:"name" json:"name"`
	Email    string `form:"email" json:"email"`
	Password string `form:"password" json:"password"`
	Role     string `form:"role" json:"role"`
	OutletId *int64 `form:"outlet_id" json:"outlet_id"`
}

func (r *UserRequest) Authorize(ctx http.Context) error {
	return nil
}

func (r *UserRequest) Filters(ctx http.Context) map[string]any {
	return map[string]any{}
}

func (r *UserRequest) Rules(ctx http.Context) map[string]any {
	rules := map[string]any{
		"name":      "required",
		"email":     "required|email",
		"role":      "required|in:owner,admin,kasir",
		"outlet_id": "integer",
	}

	if ctx.Request().Method() == stdhttp.MethodPost {
		rules["password"] = "required|min:6"
	}

	return rules
}

func (r *UserRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *UserRequest) Attributes(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *UserRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return nil
}
