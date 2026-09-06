package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/ridhoauliama97/pos-server/app/facades"
)

type M20260906000003CreateCategoriesTable struct{}

func (r *M20260906000003CreateCategoriesTable) Signature() string {
	return "20260906000003_create_categories_table"
}

func (r *M20260906000003CreateCategoriesTable) Up() error {
	if !facades.Schema().HasTable("categories") {
		if err := facades.Schema().Create("categories", func(table schema.Blueprint) {
			table.ID()
			table.String("name")
			table.UnsignedBigInteger("parent_id").Nullable()
			table.Foreign("parent_id").References("id").On("categories")
			table.Timestamps()
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260906000003CreateCategoriesTable) Down() error {
	if err := facades.Schema().DropIfExists("categories"); err != nil {
		return err
	}

	return nil
}
