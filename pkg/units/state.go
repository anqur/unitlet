package units

import (
	"context"
	"net/url"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	State interface {
		Link(ctx context.Context, loc Location) error
		Enable(ctx context.Context, name Name) error
		Disable(ctx context.Context, name Name) error
		Start(ctx context.Context, name Name) error
		Stop(ctx context.Context, name Name) error
		Reload(ctx context.Context) error
		ResetFailed(ctx context.Context, name Name) error

		List(ctx context.Context) (map[Name]*core.PodStatus, error)
		Get(ctx context.Context, name Name) (*core.PodStatus, error)
		Properties(ctx context.Context, name Name) (Properties, error)
	}

	Properties interface {
		ExitCode() int32
		RestartCount() int32
		StartedAt() meta.Time
		FinishedAt() meta.Time
		ContainerID() *url.URL
	}
)
