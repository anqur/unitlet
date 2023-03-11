package units

import (
	"context"
)

type Store interface {
	CreateUnits(ctx context.Context, us []*Unit) error
	DeleteUnit(ctx context.Context, name Name) error
	UpdateUnits(ctx context.Context, us []*Unit) error
}
