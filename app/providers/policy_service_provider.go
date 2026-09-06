package providers

import (
	"github.com/goravel/framework/contracts/foundation"

	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/policies"
)

type PolicyServiceProvider struct{}

func (r *PolicyServiceProvider) Register(app foundation.Application) {
}

func (r *PolicyServiceProvider) Boot(app foundation.Application) {
	policies.Register(facades.Gate())
}
