package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/ridhoauliama97/pos-server/app/facades"
)

type M20260906000008CreateStocksTable struct{}

func (r *M20260906000008CreateStocksTable) Signature() string {
	return "20260906000008_create_stocks_table"
}

func (r *M20260906000008CreateStocksTable) Up() error {
	if !facades.Schema().HasTable("stocks") {
		if err := facades.Schema().Create("stocks", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("outlet_id")
			table.Foreign("outlet_id").
				References("id").On("outlets")
			table.UnsignedBigInteger("product_variant_id")
			table.Foreign("product_variant_id").
				References("id").On("product_variants")
			table.Decimal("quantity").Total(10).Places(2).Default(0)
			table.Decimal("min_stock_alert").Total(10).Places(2).Default(0)
			table.Timestamps()
		}); err != nil {
			return err
		}

		if err := facades.Schema().Table("stocks", func(table schema.Blueprint) {
			table.Unique("outlet_id", "product_variant_id")
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260906000008CreateStocksTable) Down() error {
	if err := facades.Schema().DropIfExists("stocks"); err != nil {
		return err
	}

	return nil
}
