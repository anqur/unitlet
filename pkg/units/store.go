package units

import (
	"context"
)

type Store interface {
	CreateUnits(ctx context.Context, us []*Unit) error
	DeleteUnit(ctx context.Context, name UnitName) error
	UpdateUnits(ctx context.Context, us []*Unit) error
}
