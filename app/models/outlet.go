package models

import (
	"github.com/goravel/framework/database/orm"
)

type Outlet struct {
	orm.Model
	Name     string  `json:"name" db:"name"`
	Address  *string `json:"address" db:"address"`
	Phone    *string `json:"phone" db:"phone"`
	IsActive bool    `json:"is_active" db:"is_active"`
}

func (r *Outlet) TableName() string {
	return "outlets"
}
