package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/ridhoauliama97/pos-server/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20210101000001CreateJobsTable{},
		&migrations.M20260906000001CreateOutletsTable{},
		&migrations.M20260906000002CreateUsersTable{},
		&migrations.M20260906000003CreateCategoriesTable{},
		&migrations.M20260906000004CreateProductsTable{},
		&migrations.M20260906000005CreateProductVariantsTable{},
		&migrations.M20260906000006CreatePaymentMethodsTable{},
		&migrations.M20260906000007CreateCustomersTable{},
		&migrations.M20260906000008CreateStocksTable{},
		&migrations.M20260906000009CreateStockMovementsTable{},
		&migrations.M20260906000010CreateTransactionsTable{},
		&migrations.M20260906000011CreateTransactionItemsTable{},
		&migrations.M20260906000012CreateTransactionPaymentsTable{},
	}
}
