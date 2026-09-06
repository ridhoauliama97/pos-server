package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/ridhoauliama97/pos-server/app/facades"
)

type M20260906000010CreateTransactionsTable struct{}

func (r *M20260906000010CreateTransactionsTable) Signature() string {
	return "20260906000010_create_transactions_table"
}

func (r *M20260906000010CreateTransactionsTable) Up() error {
	if !facades.Schema().HasTable("transactions") {
		if err := facades.Schema().Create("transactions", func(table schema.Blueprint) {
			table.ID()
			table.String("client_uuid")
			table.Unique("client_uuid")
			table.UnsignedBigInteger("outlet_id")
			table.Foreign("outlet_id").
				References("id").On("outlets")
			table.UnsignedBigInteger("cashier_id")
			table.Foreign("cashier_id").
				References("id").On("users")
			table.UnsignedBigInteger("customer_id").Nullable()
			table.Foreign("customer_id").
				References("id").On("customers")
			table.String("invoice_number").Nullable()
			table.Decimal("subtotal").Total(10).Places(2).Default(0)
			table.Decimal("discount").Total(10).Places(2).Default(0)
			table.Decimal("tax").Total(10).Places(2).Default(0)
			table.Decimal("total").Total(10).Places(2).Default(0)
			table.Enum("status", []any{"completed", "voided", "refunded"}).Default("completed")
			table.DateTimeTz("client_created_at")
			table.DateTimeTz("created_at").UseCurrent()
			table.Index("invoice_number")
			table.Index("status")
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260906000010CreateTransactionsTable) Down() error {
	if err := facades.Schema().DropIfExists("transactions"); err != nil {
		return err
	}

	return nil
}
