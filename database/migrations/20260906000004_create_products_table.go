package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/ridhoauliama97/pos-server/app/facades"
)

type M20260906000004CreateProductsTable struct{}

func (r *M20260906000004CreateProductsTable) Signature() string {
	return "20260906000004_create_products_table"
}

func (r *M20260906000004CreateProductsTable) Up() error {
	if !facades.Schema().HasTable("products") {
		if err := facades.Schema().Create("products", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("category_id")
			table.Foreign("category_id").References("id").On("categories")
			table.String("name")
			table.String("sku").Nullable()
			table.String("barcode").Nullable()
			table.String("unit").Nullable()
			table.Boolean("is_variant").Default(false)
			table.Decimal("base_price").Total(10).Places(2).Default(0)
			table.Boolean("is_active").Default(true)
			table.Timestamps()
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260906000004CreateProductsTable) Down() error {
	if err := facades.Schema().DropIfExists("products"); err != nil {
		return err
	}

	return nil
}
