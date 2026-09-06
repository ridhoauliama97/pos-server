package v1

import (
	stdhttp "net/http"

	"github.com/goravel/framework/contracts/http"

	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/models"
	"github.com/ridhoauliama97/pos-server/app/policies"
	"github.com/ridhoauliama97/pos-server/app/services"
)

type ReportController struct{}

func NewReportController() *ReportController {
	return &ReportController{}
}

func (r *ReportController) Daily(ctx http.Context) http.Response {
	if !r.authorize(ctx, policies.AbilityReportView) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	reportService := services.NewReportService()
	report, err := reportService.Daily(
		ctx.Request().Query("from", ""),
		ctx.Request().Query("to", ""),
		r.scope(ctx),
		r.outlet(ctx),
	)
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	perDay := make([]http.Json, 0, len(report.PerDay))
	for _, day := range report.PerDay {
		perDay = append(perDay, http.Json{
			"date":              day.Date,
			"transaction_count": day.TransactionCount,
			"item_quantity":     day.ItemQuantity,
			"subtotal":          day.Subtotal,
			"discount":          day.Discount,
			"tax":               day.Tax,
			"total":             day.Total,
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"data": http.Json{
			"from":              report.From,
			"to":                report.To,
			"transaction_count": report.TransactionCount,
			"item_quantity":     report.ItemQuantity,
			"subtotal":          report.Subtotal,
			"discount":          report.Discount,
			"tax":               report.Tax,
			"total":             report.Total,
			"per_day":           perDay,
		},
	})
}

func (r *ReportController) Products(ctx http.Context) http.Response {
	if !r.authorize(ctx, policies.AbilityReportView) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	reportService := services.NewReportService()
	report, err := reportService.Products(
		ctx.Request().Query("from", ""),
		ctx.Request().Query("to", ""),
		r.scope(ctx),
		r.outlet(ctx),
	)
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	items := make([]http.Json, 0, len(report.Items))
	for _, item := range report.Items {
		items = append(items, http.Json{
			"product_variant_id": item.ProductVariantId,
			"product_name":       item.ProductName,
			"quantity":           item.Quantity,
			"subtotal":           item.Subtotal,
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"data": http.Json{
			"from":  report.From,
			"to":    report.To,
			"items": items,
		},
	})
}

func (r *ReportController) ByKasir(ctx http.Context) http.Response {
	if !r.authorize(ctx, policies.AbilityReportView) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	reportService := services.NewReportService()
	report, err := reportService.ByCashier(
		ctx.Request().Query("from", ""),
		ctx.Request().Query("to", ""),
		r.scope(ctx),
		r.outlet(ctx),
	)
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	items := make([]http.Json, 0, len(report.Items))
	for _, item := range report.Items {
		items = append(items, http.Json{
			"cashier_id":        item.CashierId,
			"cashier_name":      item.CashierName,
			"transaction_count": item.TransactionCount,
			"item_quantity":     item.ItemQuantity,
			"subtotal":          item.Subtotal,
			"discount":          item.Discount,
			"tax":               item.Tax,
			"total":             item.Total,
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"data": http.Json{
			"from":  report.From,
			"to":    report.To,
			"items": items,
		},
	})
}

func (r *ReportController) Shift(ctx http.Context) http.Response {
	if !r.authorize(ctx, policies.AbilityReportView) {
		return ctx.Response().Status(stdhttp.StatusForbidden).Json(http.Json{"error": "forbidden"})
	}

	var cashierId int64
	if r.authorize(ctx, policies.AbilityReportViewAll) {
		cashierId = int64(ctx.Request().QueryInt("kasir_id", 0))
		if cashierId <= 0 {
			return ctx.Response().Status(stdhttp.StatusUnprocessableEntity).Json(http.Json{"error": "kasir_id is required"})
		}
	} else {
		var actor models.User
		if err := facades.Auth(ctx).User(&actor); err != nil {
			return ctx.Response().Status(stdhttp.StatusUnauthorized).Json(http.Json{"error": "unauthorized"})
		}
		cashierId = int64(actor.ID)
	}

	reportService := services.NewReportService()
	report, err := reportService.Shift(
		cashierId,
		ctx.Request().Query("from", ""),
		ctx.Request().Query("to", ""),
		r.outlet(ctx),
	)
	if err != nil {
		return ctx.Response().Status(stdhttp.StatusInternalServerError).Json(http.Json{"error": err.Error()})
	}

	perDay := make([]http.Json, 0, len(report.PerDay))
	for _, day := range report.PerDay {
		perDay = append(perDay, http.Json{
			"date":              day.Date,
			"transaction_count": day.TransactionCount,
			"item_quantity":     day.ItemQuantity,
			"subtotal":          day.Subtotal,
			"discount":          day.Discount,
			"tax":               day.Tax,
			"total":             day.Total,
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"data": http.Json{
			"cashier_id":   report.CashierId,
			"cashier_name": report.CashierName,
			"from":         report.From,
			"to":           report.To,
			"per_day":      perDay,
		},
	})
}

func (r *ReportController) authorize(ctx http.Context, ability string) bool {
	return facades.Gate().WithContext(ctx).Allows(ability, nil)
}

func (r *ReportController) scope(ctx http.Context) *int64 {
	if r.authorize(ctx, policies.AbilityReportViewAll) {
		return nil
	}

	var actor models.User
	if err := facades.Auth(ctx).User(&actor); err != nil {
		return nil
	}

	kasirOnly := int64(actor.ID)

	return &kasirOnly
}

func (r *ReportController) outlet(ctx http.Context) *int64 {
	raw := ctx.Request().Query("outlet_id")
	if raw == "" {
		return nil
	}

	value := int64(ctx.Request().QueryInt("outlet_id", 0))
	if value <= 0 {
		return nil
	}

	return &value
}
