package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/ridhoauliama97/pos-server/app/facades"
)

type M20260906000011CreateTransactionItemsTable struct{}

func (r *M20260906000011CreateTransactionItemsTable) Signature() string {
	return "20260906000011_create_transaction_items_table"
}

func (r *M20260906000011CreateTransactionItemsTable) Up() error {
	if !facades.Schema().HasTable("transaction_items") {
		if err := facades.Schema().Create("transaction_items", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("transaction_id")
			table.Foreign("transaction_id").
				References("id").On("transactions")
			table.UnsignedBigInteger("product_variant_id").Nullable()
			table.Foreign("product_variant_id").
				References("id").On("product_variants")
			table.String("product_name_snapshot")
			table.Decimal("price_snapshot").Total(10).Places(2)
			table.Decimal("quantity").Total(10).Places(2)
			table.Decimal("subtotal").Total(10).Places(2)
			table.Timestamps()
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260906000011CreateTransactionItemsTable) Down() error {
	if err := facades.Schema().DropIfExists("transaction_items"); err != nil {
		return err
	}

	return nil
}
