package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/ridhoauliama97/pos-server/app/facades"
)

type M20260906000005CreateProductVariantsTable struct{}

func (r *M20260906000005CreateProductVariantsTable) Signature() string {
	return "20260906000005_create_product_variants_table"
}

func (r *M20260906000005CreateProductVariantsTable) Up() error {
	if !facades.Schema().HasTable("product_variants") {
		if err := facades.Schema().Create("product_variants", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("product_id")
			table.Foreign("product_id").References("id").On("products")
			table.String("name")
			table.String("sku_variant").Nullable()
			table.String("barcode_variant").Nullable()
			table.Decimal("price").Total(10).Places(2).Default(0)
			table.Timestamps()
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260906000005CreateProductVariantsTable) Down() error {
	if err := facades.Schema().DropIfExists("product_variants"); err != nil {
		return err
	}

	return nil
}
