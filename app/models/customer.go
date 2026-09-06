package models

import (
	"github.com/goravel/framework/database/orm"
)

type Customer struct {
	orm.Model
	Name        string  `json:"name" db:"name"`
	Phone       *string `json:"phone" db:"phone"`
	Email       *string `json:"email" db:"email"`
	MemberCode  *string `json:"member_code" db:"member_code"`
	TotalPoints int64   `json:"total_points" db:"total_points"`
}

func (r *Customer) TableName() string {
	return "customers"
}
