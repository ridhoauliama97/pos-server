package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/ridhoauliama97/pos-server/app/facades"
)

type M20260906000006CreatePaymentMethodsTable struct{}

func (r *M20260906000006CreatePaymentMethodsTable) Signature() string {
	return "20260906000006_create_payment_methods_table"
}

func (r *M20260906000006CreatePaymentMethodsTable) Up() error {
	if !facades.Schema().HasTable("payment_methods") {
		if err := facades.Schema().Create("payment_methods", func(table schema.Blueprint) {
			table.ID()
			table.String("code")
			table.Unique("code")
			table.String("name")
			table.Boolean("is_active").Default(true)
			table.Timestamps()
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260906000006CreatePaymentMethodsTable) Down() error {
	if err := facades.Schema().DropIfExists("payment_methods"); err != nil {
		return err
	}

	return nil
}
