package v1

import (
	"errors"
	stdhttp "net/http"
	"strconv"

	"github.com/goravel/framework/contracts/http"
	frmerrors "github.com/goravel/framework/errors"

	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/http/requests"
	"github.com/ridhoauliama97/pos-server/app/models"
	"github.com/ridhoauliama97/pos-server/app/policies"
)

type ProductController struct{}

func NewProductController() *ProductController {
	return &ProductController{}
}

func (r *ProductController) Index(ctx http.Context) http.Response {
	if !r.authorize(ctx, policies.AbilityProductView) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	query := facades.Orm().Query().Model(&models.Product{}).
		With("Category").With("ProductVariants").Order("id asc")
	if categoryID := ctx.Request().Query("category_id"); categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if keyword := ctx.Request().Query("q"); keyword != "" {
		query = query.Where("name ILIKE ?", "%"+keyword+"%")
	}

	var products []models.Product
	if err := query.Get(&products); err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	return ctx.Response().Success().Json(http.Json{"data": r.serializeMany(products)})
}

func (r *ProductController) Store(ctx http.Context) http.Response {
	var productRequest requests.ProductRequest
	invalid, err := ctx.Request().ValidateRequest(&productRequest)
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}
	if invalid != nil {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"errors": invalid.All()})
	}

	if !r.authorize(ctx, policies.AbilityProductCreate) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	if !r.categoryExists(productRequest.CategoryId) {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": "category not found"})
	}
	if productRequest.BasePrice < 0 {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": "base_price must not be negative"})
	}
	if err := r.validateVariants(productRequest); err != nil {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": err.Error()})
	}

	product := models.Product{
		CategoryId: productRequest.CategoryId,
		Name:       productRequest.Name,
		Sku:        productRequest.Sku,
		Barcode:    productRequest.Barcode,
		Unit:       productRequest.Unit,
		IsVariant:  productRequest.IsVariant,
		BasePrice:  productRequest.BasePrice,
		IsActive:   true,
	}

	tx, err := facades.Orm().Query().BeginTransaction()
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	if err := tx.Create(&product); err != nil {
		_ = tx.Rollback()
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	if productRequest.IsVariant {
		for _, variantRequest := range productRequest.Variants {
			variant := models.ProductVariant{
				ProductId:      int64(product.ID),
				Name:           variantRequest.Name,
				SkuVariant:     variantRequest.SkuVariant,
				BarcodeVariant: variantRequest.BarcodeVariant,
				Price:          variantRequest.Price,
			}
			if err := tx.Create(&variant); err != nil {
				_ = tx.Rollback()
				return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	return ctx.Response().Status(stdhttp.StatusCreated).Json(http.Json{"data": r.serialize(r.reload(&product))})
}

func (r *ProductController) Update(ctx http.Context) http.Response {
	var product models.Product
	if err := facades.Orm().Query().FindOrFail(&product, ctx.Request().RouteInt64("id")); err != nil {
		if errors.Is(err, frmerrors.OrmRecordNotFound) {
			return ctx.Response().Status(stdhttp.StatusNotFound).Json(http.Json{"error": "product not found"})
		}
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	var productRequest requests.ProductRequest
	invalid, err := ctx.Request().ValidateRequest(&productRequest)
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}
	if invalid != nil {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"errors": invalid.All()})
	}

	if !r.authorize(ctx, policies.AbilityProductUpdate) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	if !r.categoryExists(productRequest.CategoryId) {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": "category not found"})
	}
	if productRequest.BasePrice < 0 {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": "base_price must not be negative"})
	}
	if err := r.validateVariants(productRequest); err != nil {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": err.Error()})
	}

	product.CategoryId = productRequest.CategoryId
	product.Name = productRequest.Name
	product.Sku = productRequest.Sku
	product.Barcode = productRequest.Barcode
	product.Unit = productRequest.Unit
	product.IsVariant = productRequest.IsVariant
	product.BasePrice = productRequest.BasePrice
	if productRequest.IsActive != nil {
		product.IsActive = *productRequest.IsActive
	}

	tx, err := facades.Orm().Query().BeginTransaction()
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	if _, err := tx.Where("product_id = ?", product.ID).Delete(&models.ProductVariant{}); err != nil {
		_ = tx.Rollback()
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	if productRequest.IsVariant {
		for _, variantRequest := range productRequest.Variants {
			variant := models.ProductVariant{
				ProductId:      int64(product.ID),
				Name:           variantRequest.Name,
				SkuVariant:     variantRequest.SkuVariant,
				BarcodeVariant: variantRequest.BarcodeVariant,
				Price:          variantRequest.Price,
			}
			if err := tx.Create(&variant); err != nil {
				_ = tx.Rollback()
				return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
			}
		}
	}

	if err := tx.Save(&product); err != nil {
		_ = tx.Rollback()
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	if err := tx.Commit(); err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	return ctx.Response().Success().Json(http.Json{"data": r.serialize(r.reload(&product))})
}

func (r *ProductController) Destroy(ctx http.Context) http.Response {
	var product models.Product
	if err := facades.Orm().Query().FindOrFail(&product, ctx.Request().RouteInt64("id")); err != nil {
		if errors.Is(err, frmerrors.OrmRecordNotFound) {
			return ctx.Response().Status(stdhttp.StatusNotFound).Json(http.Json{"error": "product not found"})
		}
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	if !r.authorize(ctx, policies.AbilityProductDelete) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	product.IsActive = false
	if err := facades.Orm().Query().Save(&product); err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	return ctx.Response().Success().Json(http.Json{"data": r.serialize(product)})
}

func (r *ProductController) validateVariants(productRequest requests.ProductRequest) error {
	if !productRequest.IsVariant {
		return nil
	}
	if len(productRequest.Variants) == 0 {
		return errors.New("at least one variant is required when is_variant is true")
	}
	for i, variant := range productRequest.Variants {
		if variant.Name == "" {
			return errors.New("variants[" + strconv.Itoa(i) + "].name is required")
		}
		if variant.Price < 0 {
			return errors.New("variants[" + strconv.Itoa(i) + "].price must not be negative")
		}
	}

	return nil
}

func (r *ProductController) authorize(ctx http.Context, ability string) bool {
	return facades.Gate().WithContext(ctx).Allows(ability, nil)
}

func (r *ProductController) categoryExists(id int64) bool {
	count, err := facades.Orm().Query().Model(&models.Category{}).Where("id = ?", id).Count()
	if err != nil {
		return false
	}

	return count > 0
}

func (r *ProductController) reload(product *models.Product) models.Product {
	var refreshed models.Product
	if err := facades.Orm().Query().With("Category").With("ProductVariants").FindOrFail(&refreshed, product.ID); err != nil {
		return *product
	}

	return refreshed
}

func (r *ProductController) serialize(product models.Product) http.Json {
	variants := make([]http.Json, 0, len(product.ProductVariants))
	for _, variant := range product.ProductVariants {
		variants = append(variants, http.Json{
			"id":              variant.ID,
			"name":            variant.Name,
			"sku_variant":     variant.SkuVariant,
			"barcode_variant": variant.BarcodeVariant,
			"price":           variant.Price,
		})
	}

	result := http.Json{
		"id":          product.ID,
		"category_id": product.CategoryId,
		"category":    serializeCategory(product.Category),
		"name":        product.Name,
		"sku":         product.Sku,
		"barcode":     product.Barcode,
		"unit":        product.Unit,
		"is_variant":  product.IsVariant,
		"base_price":  product.BasePrice,
		"is_active":   product.IsActive,
		"variants":    variants,
	}

	return result
}

func (r *ProductController) serializeMany(products []models.Product) []http.Json {
	result := make([]http.Json, 0, len(products))
	for _, product := range products {
		result = append(result, r.serialize(product))
	}

	return result
}

func serializeCategory(category *models.Category) *http.Json {
	if category == nil {
		return nil
	}

	result := http.Json{
		"id":        category.ID,
		"name":      category.Name,
		"parent_id": category.ParentId,
	}

	return &result
}
