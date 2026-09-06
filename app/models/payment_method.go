package models

import (
	"github.com/goravel/framework/database/orm"
)

type PaymentMethod struct {
	orm.Model
	Code     string `json:"code" db:"code"`
	Name     string `json:"name" db:"name"`
	IsActive bool   `json:"is_active" db:"is_active"`
}

func (r *PaymentMethod) TableName() string {
	return "payment_methods"
}
