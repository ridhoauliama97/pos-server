package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"

	"github.com/ridhoauliama97/pos-server/config"
	"github.com/ridhoauliama97/pos-server/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithSeeders(Seeders).
		WithMigrations(Migrations).
		WithRouting(func() {
			routes.Api()
			routes.Web()
			routes.Grpc()
		}).
		WithProviders(Providers).
		WithConfig(config.Boot).
		Create()
}
