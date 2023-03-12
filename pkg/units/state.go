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

		Views(ctx context.Context) (Views, error)
		Properties(ctx context.Context, name Name) (Properties, error)
	}

	Views = map[string]map[string]*View
	View  struct {
		Lead   Name
		Names  []Name
		Status *core.PodStatus
	}

	Properties interface {
		ExitCode() int32
		RestartCount() int32
		StartedAt() meta.Time
		FinishedAt() meta.Time
		ContainerID() *url.URL
	}
)
