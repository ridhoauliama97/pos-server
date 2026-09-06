package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/ridhoauliama97/pos-server/app/facades"
)

type M20260906000007CreateCustomersTable struct{}

func (r *M20260906000007CreateCustomersTable) Signature() string {
	return "20260906000007_create_customers_table"
}

func (r *M20260906000007CreateCustomersTable) Up() error {
	if !facades.Schema().HasTable("customers") {
		if err := facades.Schema().Create("customers", func(table schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("phone").Nullable()
			table.String("email").Nullable()
			table.String("member_code").Nullable()
			table.Unique("member_code")
			table.UnsignedBigInteger("total_points").Default(0)
			table.Timestamps()
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260906000007CreateCustomersTable) Down() error {
	if err := facades.Schema().DropIfExists("customers"); err != nil {
		return err
	}

	return nil
}
