package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/ridhoauliama97/pos-server/app/facades"
)

type M20260906000012CreateTransactionPaymentsTable struct{}

func (r *M20260906000012CreateTransactionPaymentsTable) Signature() string {
	return "20260906000012_create_transaction_payments_table"
}

func (r *M20260906000012CreateTransactionPaymentsTable) Up() error {
	if !facades.Schema().HasTable("transaction_payments") {
		if err := facades.Schema().Create("transaction_payments", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("transaction_id")
			table.Foreign("transaction_id").
				References("id").On("transactions")
			table.UnsignedBigInteger("payment_method_id")
			table.Foreign("payment_method_id").
				References("id").On("payment_methods")
			table.Decimal("amount").Total(10).Places(2)
			table.Decimal("change_amount").Total(10).Places(2).Default(0)
			table.DateTimeTz("paid_at").UseCurrent()
			table.Timestamps()
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260906000012CreateTransactionPaymentsTable) Down() error {
	if err := facades.Schema().DropIfExists("transaction_payments"); err != nil {
		return err
	}

	return nil
}
