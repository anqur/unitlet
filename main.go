package unitlet

import (
	"github.com/virtual-kubelet/node-cli/provider"

	"github.com/anqur/unitlet/internal/states"
	"github.com/anqur/unitlet/internal/stores"
	"github.com/anqur/unitlet/pkg/providers"
	"github.com/anqur/unitlet/pkg/units"
)

const ProviderName = units.Prefix

var (
	NewFileStore = stores.NewFileStore
	NewDbusState = states.NewDbusState
)

func New(cfg provider.InitConfig) (provider.Provider, error) {
	store, err := NewFileStore(stores.DefaultFileStorePath)
	if err != nil {
		return nil, err
	}

	state, err := NewDbusState()
	if err != nil {
		return nil, err
	}

	return providers.NewUnitlet(&cfg, store, state), nil
}
