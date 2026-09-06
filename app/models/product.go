package models

import (
	"github.com/goravel/framework/database/orm"
)

type Product struct {
	orm.Model
	CategoryId int64   `json:"category_id" db:"category_id"`
	Name       string  `json:"name" db:"name"`
	Sku        *string `json:"sku" db:"sku"`
	Barcode    *string `json:"barcode" db:"barcode"`
	Unit       *string `json:"unit" db:"unit"`
	IsVariant  bool    `json:"is_variant" db:"is_variant"`
	BasePrice  float64 `json:"base_price" db:"base_price"`
	IsActive   bool    `json:"is_active" db:"is_active"`

	Category        *Category         `gorm:"foreignKey:CategoryId"`
	ProductVariants []*ProductVariant `gorm:"foreignKey:ProductId"`
}

func (r *Product) TableName() string {
	return "products"
}
