package services

import (
	"fmt"

	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/models"
)

const (
	ReportStatusCompleted = "completed"
)

type DailyReportDay struct {
	Date             string
	TransactionCount int64
	ItemQuantity     float64
	Subtotal         float64
	Discount         float64
	Tax              float64
	Total            float64
}

type DailyReport struct {
	From             string
	To               string
	TransactionCount int64
	ItemQuantity     float64
	Subtotal         float64
	Discount         float64
	Tax              float64
	Total            float64
	PerDay           []DailyReportDay
}

type ProductReportItem struct {
	ProductVariantId *int64
	ProductName      string
	Quantity         float64
	Subtotal         float64
}

type ProductReport struct {
	From  string
	To    string
	Items []ProductReportItem
}

type CashierReportItem struct {
	CashierId        int64
	CashierName      string
	TransactionCount int64
	ItemQuantity     float64
	Subtotal         float64
	Discount         float64
	Tax              float64
	Total            float64
}

type CashierReport struct {
	From  string
	To    string
	Items []CashierReportItem
}

type ShiftReportDay struct {
	Date             string
	TransactionCount int64
	ItemQuantity     float64
	Subtotal         float64
	Discount         float64
	Tax              float64
	Total            float64
}

type ShiftReport struct {
	CashierId   int64
	CashierName string
	From        string
	To          string
	PerDay      []ShiftReportDay
}

type ReportService struct{}

func NewReportService() *ReportService {
	return &ReportService{}
}

func (s *ReportService) Daily(from, to string, cashierId, outletId *int64) (*DailyReport, error) {
	transactions, err := s.loadCompleted(from, to, cashierId, outletId)
	if err != nil {
		return nil, err
	}

	report := &DailyReport{
		From: from,
		To:   to,
	}

	dayIndex := map[string]int{}
	for _, transaction := range transactions {
		report.TransactionCount++
		for _, item := range transaction.TransactionItems {
			report.ItemQuantity += item.Quantity
		}
		report.Subtotal += transaction.Subtotal
		report.Discount += transaction.Discount
		report.Tax += transaction.Tax
		report.Total += transaction.Total

		date := transactionDate(transaction)
		index, ok := dayIndex[date]
		if !ok {
			index = len(report.PerDay)
			report.PerDay = append(report.PerDay, DailyReportDay{Date: date})
			dayIndex[date] = index
		}
		day := &report.PerDay[index]
		day.TransactionCount++
		for _, item := range transaction.TransactionItems {
			day.ItemQuantity += item.Quantity
		}
		day.Subtotal += transaction.Subtotal
		day.Discount += transaction.Discount
		day.Tax += transaction.Tax
		day.Total += transaction.Total
	}

	return report, nil
}

func (s *ReportService) ByCashier(from, to string, cashierId, outletId *int64) (*CashierReport, error) {
	transactions, err := s.loadCompleted(from, to, cashierId, outletId)
	if err != nil {
		return nil, err
	}

	report := &CashierReport{
		From: from,
		To:   to,
	}

	indexByCashier := map[int64]int{}
	for _, transaction := range transactions {
		index, ok := indexByCashier[transaction.CashierId]
		if !ok {
			name := cashierName(transaction)
			report.Items = append(report.Items, CashierReportItem{
				CashierId:   transaction.CashierId,
				CashierName: name,
			})
			index = len(report.Items) - 1
			indexByCashier[transaction.CashierId] = index
		}
		row := &report.Items[index]
		row.TransactionCount++
		for _, item := range transaction.TransactionItems {
			row.ItemQuantity += item.Quantity
		}
		row.Subtotal += transaction.Subtotal
		row.Discount += transaction.Discount
		row.Tax += transaction.Tax
		row.Total += transaction.Total
	}

	return report, nil
}

func (s *ReportService) Shift(cashierId int64, from, to string, outletId *int64) (*ShiftReport, error) {
	transactions, err := s.loadCompleted(from, to, &cashierId, outletId)
	if err != nil {
		return nil, err
	}

	report := &ShiftReport{
		CashierId: cashierId,
		From:      from,
		To:        to,
	}
	if len(transactions) > 0 {
		report.CashierName = cashierName(transactions[0])
	}
	if report.CashierName == "" {
		var cashier models.User
		if err := facades.Orm().Query().Where("id = ?", cashierId).FirstOrFail(&cashier); err == nil {
			report.CashierName = cashier.Name
		}
	}

	dayIndex := map[string]int{}
	for _, transaction := range transactions {
		date := transactionDate(transaction)
		index, ok := dayIndex[date]
		if !ok {
			index = len(report.PerDay)
			report.PerDay = append(report.PerDay, ShiftReportDay{Date: date})
			dayIndex[date] = index
		}
		day := &report.PerDay[index]
		day.TransactionCount++
		for _, item := range transaction.TransactionItems {
			day.ItemQuantity += item.Quantity
		}
		day.Subtotal += transaction.Subtotal
		day.Discount += transaction.Discount
		day.Tax += transaction.Tax
		day.Total += transaction.Total
	}

	return report, nil
}

func (s *ReportService) Products(from, to string, cashierId, outletId *int64) (*ProductReport, error) {
	transactions, err := s.loadCompleted(from, to, cashierId, outletId)
	if err != nil {
		return nil, err
	}

	report := &ProductReport{
		From: from,
		To:   to,
	}

	itemIndex := map[string]int{}
	for _, transaction := range transactions {
		for _, item := range transaction.TransactionItems {
			key := productKey(item)
			index, ok := itemIndex[key]
			if !ok {
				index = len(report.Items)
				report.Items = append(report.Items, ProductReportItem{
					ProductVariantId: item.ProductVariantId,
					ProductName:      item.ProductNameSnapshot,
				})
				itemIndex[key] = index
			}
			row := &report.Items[index]
			row.Quantity += item.Quantity
			row.Subtotal += item.Subtotal
		}
	}

	return report, nil
}

func (s *ReportService) loadCompleted(from, to string, cashierId, outletId *int64) ([]*models.Transaction, error) {
	query := facades.Orm().Query().
		Model(&models.Transaction{}).
		With("TransactionItems").
		With("Cashier").
		Where("status = ?", ReportStatusCompleted)
	if from != "" {
		query = query.Where("created_at >= ?", fmt.Sprintf("%s 00:00:00", from))
	}
	if to != "" {
		query = query.Where("created_at <= ?", fmt.Sprintf("%s 23:59:59", to))
	}
	if outletId != nil {
		query = query.Where("outlet_id = ?", *outletId)
	}
	if cashierId != nil {
		query = query.Where("cashier_id = ?", *cashierId)
	}

	var transactions []*models.Transaction
	if err := query.Order("created_at asc").Get(&transactions); err != nil {
		return nil, err
	}

	return transactions, nil
}

func transactionDate(transaction *models.Transaction) string {
	if transaction.CreatedAt.Carbon == nil {
		return ""
	}

	return transaction.CreatedAt.Carbon.ToDateString()
}

func productKey(item *models.TransactionItem) string {
	if item.ProductVariantId == nil {
		return fmt.Sprintf("0|%s", item.ProductNameSnapshot)
	}

	return fmt.Sprintf("%d|%s", *item.ProductVariantId, item.ProductNameSnapshot)
}

func cashierName(transaction *models.Transaction) string {
	if transaction.Cashier != nil {
		return transaction.Cashier.Name
	}

	return ""
}
