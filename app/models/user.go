package models

import (
	"github.com/goravel/framework/database/orm"
)

const (
	RoleOwner = "owner"
	RoleAdmin = "admin"
	RoleKasir = "kasir"
)

type User struct {
	orm.Model
	OutletId *int64 `json:"outlet_id" db:"outlet_id"`
	Name     string `json:"name" db:"name"`
	Email    string `json:"email" db:"email"`
	Password string `json:"-" db:"password"`
	Role     string `json:"role" db:"role"`
	IsActive bool   `json:"is_active" db:"is_active"`

	Outlet *Outlet `gorm:"foreignKey:OutletId"`
}

func (r *User) TableName() string {
	return "users"
}
