package models

import (
	"github.com/goravel/framework/database/orm"
)

type ProductVariant struct {
	orm.Model
	ProductId      int64   `json:"product_id" db:"product_id"`
	Name           string  `json:"name" db:"name"`
	SkuVariant     *string `json:"sku_variant" db:"sku_variant"`
	BarcodeVariant *string `json:"barcode_variant" db:"barcode_variant"`
	Price          float64 `json:"price" db:"price"`

	Product *Product `gorm:"foreignKey:ProductId"`
}

func (r *ProductVariant) TableName() string {
	return "product_variants"
}
