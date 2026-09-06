package v1

import (
	"errors"
	"fmt"
	stdhttp "net/http"

	"github.com/goravel/framework/contracts/http"
	frmerrors "github.com/goravel/framework/errors"
	"math/rand/v2"

	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/http/requests"
	"github.com/ridhoauliama97/pos-server/app/models"
	"github.com/ridhoauliama97/pos-server/app/policies"
	"github.com/ridhoauliama97/pos-server/app/services"
)

type CustomerController struct{}

func NewCustomerController() *CustomerController {
	return &CustomerController{}
}

func (r *CustomerController) Index(ctx http.Context) http.Response {
	if !r.authorize(ctx, policies.AbilityCustomerView) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	query := facades.Orm().Query().Model(&models.Customer{}).Order("id asc")
	if keyword := ctx.Request().Query("q"); keyword != "" {
		query = query.Where("name ILIKE ? OR phone ILIKE ? OR member_code ILIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	var customers []models.Customer
	if err := query.Get(&customers); err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	return ctx.Response().Success().Json(http.Json{"data": r.serializeMany(customers)})
}

func (r *CustomerController) Store(ctx http.Context) http.Response {
	var customerRequest requests.CustomerRequest
	invalid, err := ctx.Request().ValidateRequest(&customerRequest)
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}
	if invalid != nil {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"errors": invalid.All()})
	}

	if !r.authorize(ctx, policies.AbilityCustomerCreate) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	customer := models.Customer{
		Name:        customerRequest.Name,
		Phone:       customerRequest.Phone,
		Email:       customerRequest.Email,
		TotalPoints: 0,
	}
	if customerRequest.MemberCode != nil && *customerRequest.MemberCode != "" {
		code := *customerRequest.MemberCode
		if r.memberCodeExists(code) {
			return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": "member_code already exists"})
		}
		customer.MemberCode = &code
	} else {
		code, err := r.generateMemberCode()
		if err != nil {
			return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
		}
		customer.MemberCode = &code
	}

	if err := facades.Orm().Query().Create(&customer); err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	return ctx.Response().Status(stdhttp.StatusCreated).Json(http.Json{"data": r.serialize(customer)})
}

func (r *CustomerController) Show(ctx http.Context) http.Response {
	if !r.authorize(ctx, policies.AbilityCustomerView) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	customer, status := r.find(ctx)
	if customer == nil {
		return ctx.Response().Status(status).Json(http.Json{"error": "customer not found"})
	}

	return ctx.Response().Success().Json(http.Json{"data": r.serialize(*customer)})
}

func (r *CustomerController) Update(ctx http.Context) http.Response {
	customer, status := r.find(ctx)
	if customer == nil {
		return ctx.Response().Status(status).Json(http.Json{"error": "customer not found"})
	}

	var customerRequest requests.CustomerRequest
	invalid, err := ctx.Request().ValidateRequest(&customerRequest)
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}
	if invalid != nil {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"errors": invalid.All()})
	}

	if !r.authorize(ctx, policies.AbilityCustomerUpdate) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	customer.Name = customerRequest.Name
	if customerRequest.Phone != nil {
		customer.Phone = customerRequest.Phone
	}
	if customerRequest.Email != nil {
		customer.Email = customerRequest.Email
	}
	if customerRequest.MemberCode != nil && *customerRequest.MemberCode != "" {
		code := *customerRequest.MemberCode
		if r.memberCodeExistsForOther(code, int64(customer.ID)) {
			return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": "member_code already exists"})
		}
		customer.MemberCode = &code
	}

	if err := facades.Orm().Query().Save(customer); err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	return ctx.Response().Success().Json(http.Json{"data": r.serialize(*customer)})
}

func (r *CustomerController) Destroy(ctx http.Context) http.Response {
	customer, status := r.find(ctx)
	if customer == nil {
		return ctx.Response().Status(status).Json(http.Json{"error": "customer not found"})
	}

	if !r.authorize(ctx, policies.AbilityCustomerDelete) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	references, err := facades.Orm().Query().Model(&models.Transaction{}).Where("customer_id = ?", customer.ID).Count()
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}
	if references > 0 {
		return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": "customer has transactions"})
	}

	if _, err := facades.Orm().Query().Delete(customer); err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	return ctx.Response().Success().Json(http.Json{"message": "customer deleted"})
}

func (r *CustomerController) Transactions(ctx http.Context) http.Response {
	if !r.authorize(ctx, policies.AbilityCustomerView) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	customer, status := r.find(ctx)
	if customer == nil {
		return ctx.Response().Status(status).Json(http.Json{"error": "customer not found"})
	}

	page := ctx.Request().QueryInt("page", 1)
	limit := ctx.Request().QueryInt("limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	service := services.NewCustomerService()
	transactions, total, err := service.History(int64(customer.ID), page, limit)
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	items := make([]http.Json, 0, len(transactions))
	for _, transaction := range transactions {
		items = append(items, r.serializeHistory(transaction))
	}

	return ctx.Response().Success().Json(http.Json{
		"data": items,
		"meta": http.Json{
			"customer_id": customer.ID,
			"page":        page,
			"limit":       limit,
			"total":       total,
		},
	})
}

func (r *CustomerController) serializeHistory(transaction *models.Transaction) http.Json {
	items := make([]http.Json, 0, len(transaction.TransactionItems))
	for _, item := range transaction.TransactionItems {
		items = append(items, http.Json{
			"product_variant_id":    item.ProductVariantId,
			"product_name_snapshot": item.ProductNameSnapshot,
			"price_snapshot":        item.PriceSnapshot,
			"quantity":              item.Quantity,
			"subtotal":              item.Subtotal,
		})
	}

	return http.Json{
		"id":                transaction.Id,
		"invoice_number":    transaction.InvoiceNumber,
		"outlet_id":         transaction.OutletId,
		"cashier_id":        transaction.CashierId,
		"status":            transaction.Status,
		"subtotal":          transaction.Subtotal,
		"discount":          transaction.Discount,
		"tax":               transaction.Tax,
		"total":             transaction.Total,
		"client_created_at": transaction.ClientCreatedAt,
		"created_at":        transaction.CreatedAt,
		"items":             items,
	}
}

func (r *CustomerController) find(ctx http.Context) (*models.Customer, int) {
	var customer models.Customer
	if err := facades.Orm().Query().FindOrFail(&customer, ctx.Request().RouteInt64("id")); err != nil {
		if errors.Is(err, frmerrors.OrmRecordNotFound) {
			return nil, stdhttp.StatusNotFound
		}
		return nil, stdhttp.StatusInternalServerError
	}

	return &customer, stdhttp.StatusOK
}

func (r *CustomerController) authorize(ctx http.Context, ability string) bool {
	return facades.Gate().WithContext(ctx).Allows(ability, nil)
}

func (r *CustomerController) memberCodeExists(code string) bool {
	count, err := facades.Orm().Query().Model(&models.Customer{}).Where("member_code = ?", code).Count()
	if err != nil {
		return true
	}

	return count > 0
}

func (r *CustomerController) memberCodeExistsForOther(code string, customerId int64) bool {
	count, err := facades.Orm().Query().Model(&models.Customer{}).
		Where("member_code = ? AND id != ?", code, customerId).Count()
	if err != nil {
		return true
	}

	return count > 0
}

func (r *CustomerController) generateMemberCode() (string, error) {
	for i := 0; i < 5; i++ {
		code := fmt.Sprintf("MBR-%06d", rand.IntN(1000000))
		if !r.memberCodeExists(code) {
			return code, nil
		}
	}

	return "", errors.New("failed to generate unique member_code")
}

func (r *CustomerController) serialize(customer models.Customer) http.Json {
	return http.Json{
		"id":           customer.ID,
		"name":         customer.Name,
		"phone":        customer.Phone,
		"email":        customer.Email,
		"member_code":  customer.MemberCode,
		"total_points": customer.TotalPoints,
	}
}

func (r *CustomerController) serializeMany(customers []models.Customer) []http.Json {
	result := make([]http.Json, 0, len(customers))
	for _, customer := range customers {
		result = append(result, r.serialize(customer))
	}

	return result
}
