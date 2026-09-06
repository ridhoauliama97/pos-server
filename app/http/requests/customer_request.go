package requests

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type CustomerRequest struct {
	Name       string  `form:"name" json:"name"`
	Phone      *string `form:"phone" json:"phone"`
	Email      *string `form:"email" json:"email"`
	MemberCode *string `form:"member_code" json:"member_code"`
}

func (r *CustomerRequest) Authorize(ctx http.Context) error {
	return nil
}

func (r *CustomerRequest) Filters(ctx http.Context) map[string]any {
	return map[string]any{}
}

func (r *CustomerRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":        "required",
		"phone":       "string",
		"email":       "email",
		"member_code": "string",
	}
}

func (r *CustomerRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *CustomerRequest) Attributes(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *CustomerRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return nil
}
