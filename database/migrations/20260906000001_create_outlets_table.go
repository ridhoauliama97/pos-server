package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/ridhoauliama97/pos-server/app/facades"
)

type M20260906000001CreateOutletsTable struct{}

func (r *M20260906000001CreateOutletsTable) Signature() string {
	return "20260906000001_create_outlets_table"
}

func (r *M20260906000001CreateOutletsTable) Up() error {
	if !facades.Schema().HasTable("outlets") {
		if err := facades.Schema().Create("outlets", func(table schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("address").Nullable()
			table.String("phone").Nullable()
			table.Boolean("is_active").Default(true)
			table.Timestamps()
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260906000001CreateOutletsTable) Down() error {
	if err := facades.Schema().DropIfExists("outlets"); err != nil {
		return err
	}

	return nil
}
