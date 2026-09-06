package v1

import (
	"errors"
	stdhttp "net/http"

	"github.com/goravel/framework/contracts/http"
	frmerrors "github.com/goravel/framework/errors"

	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/http/requests"
	"github.com/ridhoauliama97/pos-server/app/models"
	"github.com/ridhoauliama97/pos-server/app/policies"
)

type UserController struct{}

func NewUserController() *UserController {
	return &UserController{}
}

func (r *UserController) Index(ctx http.Context) http.Response {
	if !r.authorize(ctx, policies.AbilityUserView, nil) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	var actor models.User
	if err := facades.Auth(ctx).User(&actor); err != nil {
		return ctx.Response().Status(stdhttp.StatusUnauthorized).Json(http.Json{"error": "unauthorized"})
	}

	query := facades.Orm().Query().Model(&models.User{}).Order("id asc")
	if actor.Role == models.RoleAdmin {
		query = query.Where("role != ?", models.RoleOwner)
	}

	var users []models.User
	if err := query.Get(&users); err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	return ctx.Response().Success().Json(http.Json{"data": r.serializeMany(users)})
}

func (r *UserController) Store(ctx http.Context) http.Response {
	var userRequest requests.UserRequest
	invalid, err := ctx.Request().ValidateRequest(&userRequest)
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}
	if invalid != nil {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"errors": invalid.All()})
	}

	if !r.authorize(ctx, policies.AbilityUserCreate, map[string]any{"role": userRequest.Role}) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	count, err := facades.Orm().Query().Model(&models.User{}).Where("email", userRequest.Email).Count()
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}
	if count > 0 {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": "email already taken"})
	}

	password, err := facades.Hash().Make(userRequest.Password)
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	user := models.User{
		Name:     userRequest.Name,
		Email:    userRequest.Email,
		Password: password,
		Role:     userRequest.Role,
		IsActive: true,
		OutletId: userRequest.OutletId,
	}
	if err := facades.Orm().Query().Create(&user); err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	return ctx.Response().Status(stdhttp.StatusCreated).Json(http.Json{"data": r.serialize(user)})
}

func (r *UserController) Update(ctx http.Context) http.Response {
	var user models.User
	if err := facades.Orm().Query().FindOrFail(&user, ctx.Request().RouteInt64("id")); err != nil {
		if isModelNotFound(err) {
			return ctx.Response().Status(stdhttp.StatusNotFound).Json(http.Json{"error": "user not found"})
		}
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	var userRequest requests.UserRequest
	invalid, err := ctx.Request().ValidateRequest(&userRequest)
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}
	if invalid != nil {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"errors": invalid.All()})
	}

	if userRequest.Password != "" && len(userRequest.Password) < 6 {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": "password must be at least 6 characters"})
	}

	targetRole := user.Role
	if userRequest.Role != "" {
		targetRole = userRequest.Role
	}

	if !r.authorize(ctx, policies.AbilityUserUpdate, map[string]any{"role": targetRole}) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	user.Name = userRequest.Name
	user.Email = userRequest.Email
	user.Role = targetRole
	if userRequest.OutletId != nil {
		user.OutletId = userRequest.OutletId
	}
	if userRequest.Password != "" {
		hashed, err := facades.Hash().Make(userRequest.Password)
		if err != nil {
			return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
		}
		user.Password = hashed
	}

	if err := facades.Orm().Query().Save(&user); err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	return ctx.Response().Success().Json(http.Json{"data": r.serialize(user)})
}

func (r *UserController) Destroy(ctx http.Context) http.Response {
	var user models.User
	if err := facades.Orm().Query().FindOrFail(&user, ctx.Request().RouteInt64("id")); err != nil {
		if isModelNotFound(err) {
			return ctx.Response().Status(stdhttp.StatusNotFound).Json(http.Json{"error": "user not found"})
		}
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	var actor models.User
	if err := facades.Auth(ctx).User(&actor); err != nil {
		return ctx.Response().Status(stdhttp.StatusUnauthorized).Json(http.Json{"error": "unauthorized"})
	}
	if actor.ID == user.ID {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": "cannot deactivate yourself"})
	}

	if !r.authorize(ctx, policies.AbilityUserDelete, map[string]any{"role": user.Role}) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	user.IsActive = false
	if err := facades.Orm().Query().Save(&user); err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	return ctx.Response().Success().Json(http.Json{"data": r.serialize(user)})
}

func (r *UserController) authorize(ctx http.Context, ability string, arguments map[string]any) bool {
	return facades.Gate().WithContext(ctx).Allows(ability, arguments)
}

func (r *UserController) serialize(user models.User) http.Json {
	return http.Json{
		"id":        user.ID,
		"name":      user.Name,
		"email":     user.Email,
		"role":      user.Role,
		"is_active": user.IsActive,
		"outlet_id": user.OutletId,
	}
}

func (r *UserController) serializeMany(users []models.User) []http.Json {
	result := make([]http.Json, 0, len(users))
	for _, user := range users {
		result = append(result, r.serialize(user))
	}

	return result
}

func isModelNotFound(err error) bool {
	return err != nil && errors.Is(err, frmerrors.OrmRecordNotFound)
}
