package units

import (
	"context"
	"net/url"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	State interface {
		Enable(ctx context.Context, name UnitName) error
		Disable(ctx context.Context, name UnitName) error
		Start(ctx context.Context, name UnitName) error
		Stop(ctx context.Context, name UnitName) error
		Reload(ctx context.Context) error
		ResetFailed(ctx context.Context, name UnitName) error

		List(ctx context.Context) (map[UnitName]*core.Pod, error)
		Get(ctx context.Context, name UnitName) (*core.Pod, error)
		Properties(ctx context.Context, name UnitName) (Properties, error)
	}

	Properties interface {
		ExitCode() int32
		StartedAt() meta.Time
		FinishedAt() meta.Time
		ContainerID() *url.URL
	}
)
