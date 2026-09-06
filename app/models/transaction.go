package models

import (
	"github.com/goravel/framework/support/carbon"
)

type Transaction struct {
	Id              int64           `json:"id" db:"id" gorm:"primaryKey"`
	ClientUuid      string          `json:"client_uuid" db:"client_uuid"`
	OutletId        int64           `json:"outlet_id" db:"outlet_id"`
	CashierId       int64           `json:"cashier_id" db:"cashier_id"`
	CustomerId      *int64          `json:"customer_id" db:"customer_id"`
	InvoiceNumber   *string         `json:"invoice_number" db:"invoice_number"`
	Subtotal        float64         `json:"subtotal" db:"subtotal"`
	Discount        float64         `json:"discount" db:"discount"`
	Tax             float64         `json:"tax" db:"tax"`
	Total           float64         `json:"total" db:"total"`
	Status          string          `json:"status" db:"status"`
	ClientCreatedAt carbon.DateTime `json:"client_created_at" db:"client_created_at"`
	CreatedAt       carbon.DateTime `json:"created_at" db:"created_at"`

	Outlet              *Outlet               `gorm:"foreignKey:OutletId"`
	Cashier             *User                 `gorm:"foreignKey:CashierId"`
	Customer            *Customer             `gorm:"foreignKey:CustomerId"`
	TransactionItems    []*TransactionItem    `gorm:"foreignKey:TransactionId"`
	TransactionPayments []*TransactionPayment `gorm:"foreignKey:TransactionId"`
}

func (r *Transaction) TableName() string {
	return "transactions"
}
