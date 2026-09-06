package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/ridhoauliama97/pos-server/app/facades"
)

type M20260906000009CreateStockMovementsTable struct{}

func (r *M20260906000009CreateStockMovementsTable) Signature() string {
	return "20260906000009_create_stock_movements_table"
}

func (r *M20260906000009CreateStockMovementsTable) Up() error {
	if !facades.Schema().HasTable("stock_movements") {
		if err := facades.Schema().Create("stock_movements", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("outlet_id")
			table.Foreign("outlet_id").
				References("id").On("outlets")
			table.UnsignedBigInteger("product_variant_id")
			table.Foreign("product_variant_id").
				References("id").On("product_variants")
			table.Enum("type", []any{"in", "out", "adjustment", "sale", "return"})
			table.Decimal("quantity").Total(10).Places(2)
			table.String("reference_type").Nullable()
			table.UnsignedBigInteger("reference_id").Nullable()
			table.String("note").Nullable()
			table.UnsignedBigInteger("created_by").Nullable()
			table.Foreign("created_by").
				References("id").On("users")
			table.DateTimeTz("created_at").UseCurrent()
			table.Index("reference_type", "reference_id")
			table.Index("created_at")
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260906000009CreateStockMovementsTable) Down() error {
	if err := facades.Schema().DropIfExists("stock_movements"); err != nil {
		return err
	}

	return nil
}
