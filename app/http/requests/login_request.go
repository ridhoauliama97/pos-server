package requests

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type LoginRequest struct {
	Email    string `form:"email" json:"email"`
	Password string `form:"password" json:"password"`
}

func (r *LoginRequest) Authorize(ctx http.Context) error {
	return nil
}

func (r *LoginRequest) Filters(ctx http.Context) map[string]any {
	return map[string]any{}
}

func (r *LoginRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"email":    "required|email",
		"password": "required",
	}
}

func (r *LoginRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *LoginRequest) Attributes(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *LoginRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return nil
}
