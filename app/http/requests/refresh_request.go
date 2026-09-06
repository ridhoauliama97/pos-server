package requests

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type RefreshRequest struct {
}

func (r *RefreshRequest) Authorize(ctx http.Context) error {
	return nil
}

func (r *RefreshRequest) Filters(ctx http.Context) map[string]any {
	return map[string]any{}
}

func (r *RefreshRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{}
}

func (r *RefreshRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *RefreshRequest) Attributes(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *RefreshRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return nil
}
