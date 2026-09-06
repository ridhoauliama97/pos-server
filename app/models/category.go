package models

import (
	"github.com/goravel/framework/database/orm"
)

type Category struct {
	orm.Model
	Name     string `json:"name" db:"name"`
	ParentId *int64 `json:"parent_id" db:"parent_id"`

	Parent   *Category   `gorm:"foreignKey:ParentId"`
	Children []*Category `gorm:"foreignKey:ParentId"`
}

func (r *Category) TableName() string {
	return "categories"
}
