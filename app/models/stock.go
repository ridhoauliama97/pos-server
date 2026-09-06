package models

import (
	"github.com/goravel/framework/database/orm"
)

type Stock struct {
	orm.Model
	OutletId         int64   `json:"outlet_id" db:"outlet_id"`
	ProductVariantId int64   `json:"product_variant_id" db:"product_variant_id"`
	Quantity         float64 `json:"quantity" db:"quantity"`
	MinStockAlert    float64 `json:"min_stock_alert" db:"min_stock_alert"`

	Outlet         *Outlet         `gorm:"foreignKey:OutletId"`
	ProductVariant *ProductVariant `gorm:"foreignKey:ProductVariantId"`
}

func (r *Stock) TableName() string {
	return "stocks"
}
