package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/ridhoauliama97/pos-server/app/facades"
)

type M20260906000002CreateUsersTable struct{}

func (r *M20260906000002CreateUsersTable) Signature() string {
	return "20260906000002_create_users_table"
}

func (r *M20260906000002CreateUsersTable) Up() error {
	if !facades.Schema().HasTable("users") {
		if err := facades.Schema().Create("users", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("outlet_id").Nullable()
			table.Foreign("outlet_id").References("id").On("outlets")
			table.String("name")
			table.String("email")
			table.Unique("email")
			table.String("password")
			table.Enum("role", []any{"owner", "admin", "kasir"}).Default("kasir")
			table.Boolean("is_active").Default(true)
			table.Timestamps()
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260906000002CreateUsersTable) Down() error {
	if err := facades.Schema().DropIfExists("users"); err != nil {
		return err
	}

	return nil
}
