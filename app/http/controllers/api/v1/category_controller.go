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

type CategoryController struct{}

func NewCategoryController() *CategoryController {
	return &CategoryController{}
}

func (r *CategoryController) Index(ctx http.Context) http.Response {
	if !r.authorize(ctx, policies.AbilityCategoryView) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	query := facades.Orm().Query().Model(&models.Category{}).Order("id asc")
	if parentID := ctx.Request().Query("parent_id"); parentID != "" {
		query = query.Where("parent_id = ?", parentID)
	}

	var categories []models.Category
	if err := query.Get(&categories); err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	return ctx.Response().Success().Json(http.Json{"data": r.serializeMany(categories)})
}

func (r *CategoryController) Store(ctx http.Context) http.Response {
	var categoryRequest requests.CategoryRequest
	invalid, err := ctx.Request().ValidateRequest(&categoryRequest)
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}
	if invalid != nil {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"errors": invalid.All()})
	}

	if !r.authorize(ctx, policies.AbilityCategoryCreate) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	if categoryRequest.ParentId != nil && !r.parentExists(*categoryRequest.ParentId) {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": "parent category not found"})
	}

	category := models.Category{
		Name:     categoryRequest.Name,
		ParentId: categoryRequest.ParentId,
	}
	if err := facades.Orm().Query().Create(&category); err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	return ctx.Response().Status(stdhttp.StatusCreated).Json(http.Json{"data": r.serialize(category)})
}

func (r *CategoryController) Update(ctx http.Context) http.Response {
	var category models.Category
	if err := facades.Orm().Query().FindOrFail(&category, ctx.Request().RouteInt64("id")); err != nil {
		if errors.Is(err, frmerrors.OrmRecordNotFound) {
			return ctx.Response().Status(stdhttp.StatusNotFound).Json(http.Json{"error": "category not found"})
		}
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	var categoryRequest requests.CategoryRequest
	invalid, err := ctx.Request().ValidateRequest(&categoryRequest)
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}
	if invalid != nil {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"errors": invalid.All()})
	}

	if !r.authorize(ctx, policies.AbilityCategoryUpdate) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	category.Name = categoryRequest.Name
	if categoryRequest.ParentId != nil {
		if int64(category.ID) == *categoryRequest.ParentId || !r.parentExists(*categoryRequest.ParentId) {
			return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": "invalid parent category"})
		}
		category.ParentId = categoryRequest.ParentId
	}

	if err := facades.Orm().Query().Save(&category); err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	return ctx.Response().Success().Json(http.Json{"data": r.serialize(category)})
}

func (r *CategoryController) Destroy(ctx http.Context) http.Response {
	var category models.Category
	if err := facades.Orm().Query().FindOrFail(&category, ctx.Request().RouteInt64("id")); err != nil {
		if errors.Is(err, frmerrors.OrmRecordNotFound) {
			return ctx.Response().Status(stdhttp.StatusNotFound).Json(http.Json{"error": "category not found"})
		}
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	if !r.authorize(ctx, policies.AbilityCategoryDelete) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	children, err := facades.Orm().Query().Model(&models.Category{}).Where("parent_id = ?", category.ID).Count()
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}
	if children > 0 {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": "category has sub-categories"})
	}

	products, err := facades.Orm().Query().Model(&models.Product{}).Where("category_id = ?", category.ID).Count()
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}
	if products > 0 {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": "category has products"})
	}

	if _, err := facades.Orm().Query().Delete(&category); err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	return ctx.Response().Success().Json(http.Json{"message": "category deleted"})
}

func (r *CategoryController) authorize(ctx http.Context, ability string) bool {
	return facades.Gate().WithContext(ctx).Allows(ability, nil)
}

func (r *CategoryController) parentExists(id int64) bool {
	count, err := facades.Orm().Query().Model(&models.Category{}).Where("id = ?", id).Count()
	if err != nil {
		return false
	}

	return count > 0
}

func (r *CategoryController) serialize(category models.Category) http.Json {
	return http.Json{
		"id":        category.ID,
		"name":      category.Name,
		"parent_id": category.ParentId,
	}
}

func (r *CategoryController) serializeMany(categories []models.Category) []http.Json {
	result := make([]http.Json, 0, len(categories))
	for _, category := range categories {
		result = append(result, r.serialize(category))
	}

	return result
}
