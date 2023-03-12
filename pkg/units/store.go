package units

import (
	"context"
)

type (
	Location string

	Store interface {
		Location(name Name) Location
		GetUnit(ctx context.Context, name Name) (*Unit, error)
		CreateUnits(ctx context.Context, us []*Unit) error
		DeleteUnit(ctx context.Context, name Name) error
		UpdateUnits(ctx context.Context, us []*Unit) error
	}
)
