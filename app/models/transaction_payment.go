package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/carbon"
)

type TransactionPayment struct {
	orm.Model
	TransactionId   int64           `json:"transaction_id" db:"transaction_id"`
	PaymentMethodId int64           `json:"payment_method_id" db:"payment_method_id"`
	Amount          float64         `json:"amount" db:"amount"`
	ChangeAmount    float64         `json:"change_amount" db:"change_amount"`
	PaidAt          carbon.DateTime `json:"paid_at" db:"paid_at"`

	Transaction   *Transaction   `gorm:"foreignKey:TransactionId"`
	PaymentMethod *PaymentMethod `gorm:"foreignKey:PaymentMethodId"`
}

func (r *TransactionPayment) TableName() string {
	return "transaction_payments"
}
