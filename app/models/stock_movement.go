package models

import (
	"github.com/goravel/framework/support/carbon"
)

type StockMovement struct {
	Id               int64           `json:"id" db:"id" gorm:"primaryKey"`
	OutletId         int64           `json:"outlet_id" db:"outlet_id"`
	ProductVariantId int64           `json:"product_variant_id" db:"product_variant_id"`
	Type             string          `json:"type" db:"type"`
	Quantity         float64         `json:"quantity" db:"quantity"`
	ReferenceType    *string         `json:"reference_type" db:"reference_type"`
	ReferenceId      *int64          `json:"reference_id" db:"reference_id"`
	Note             *string         `json:"note" db:"note"`
	CreatedBy        *int64          `json:"created_by" db:"created_by"`
	CreatedAt        carbon.DateTime `json:"created_at" db:"created_at"`

	Outlet         *Outlet         `gorm:"foreignKey:OutletId"`
	ProductVariant *ProductVariant `gorm:"foreignKey:ProductVariantId"`
	CreatedByUser  *User           `gorm:"foreignKey:CreatedBy"`
}

func (r *StockMovement) TableName() string {
	return "stock_movements"
}
