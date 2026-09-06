package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/ridhoauliama97/pos-server/database/seeders"
)

func Seeders() []seeder.Seeder {
	return []seeder.Seeder{
		&seeders.DatabaseSeeder{},
	}
}
