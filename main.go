package unitlet

import (
	"github.com/virtual-kubelet/node-cli/provider"

	"github.com/anqur/unitlet/internal/states"
	"github.com/anqur/unitlet/internal/stores"
	"github.com/anqur/unitlet/pkg/providers"
	"github.com/anqur/unitlet/pkg/units"
)

const ProviderName = units.UnitPrefix

var (
	NewFileStore = stores.NewFileStore
	NewDbusState = states.NewDbusState
)

func New(provider.InitConfig) (provider.Provider, error) {
	store, err := NewFileStore(stores.DefaultUnitFileStorePath)
	if err != nil {
		return nil, err
	}

	state, err := NewDbusState()
	if err != nil {
		return nil, err
	}

	return providers.NewUnitlet(store, state), nil
}
