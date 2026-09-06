package models

import (
	"github.com/goravel/framework/database/orm"
)

type TransactionItem struct {
	orm.Model
	TransactionId       int64   `json:"transaction_id" db:"transaction_id"`
	ProductVariantId    *int64  `json:"product_variant_id" db:"product_variant_id"`
	ProductNameSnapshot string  `json:"product_name_snapshot" db:"product_name_snapshot"`
	PriceSnapshot       float64 `json:"price_snapshot" db:"price_snapshot"`
	Quantity            float64 `json:"quantity" db:"quantity"`
	Subtotal            float64 `json:"subtotal" db:"subtotal"`

	Transaction    *Transaction    `gorm:"foreignKey:TransactionId"`
	ProductVariant *ProductVariant `gorm:"foreignKey:ProductVariantId"`
}

func (r *TransactionItem) TableName() string {
	return "transaction_items"
}
