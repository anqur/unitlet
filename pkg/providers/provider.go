package providers

import (
	"context"
	"io"
	"sync"

	"github.com/virtual-kubelet/node-cli/provider"
	"github.com/virtual-kubelet/virtual-kubelet/node/api"
	v1 "k8s.io/api/core/v1"

	"github.com/anqur/unitlet/pkg/errs"
	"github.com/anqur/unitlet/pkg/units"
)

type Unitlet struct {
	mu    sync.RWMutex
	cfg   *provider.InitConfig
	store units.Store
	state units.State
}

func NewUnitlet(cfg *provider.InitConfig, store units.Store, state units.State) provider.Provider {
	return &Unitlet{cfg: cfg, store: store, state: state}
}

func (l *Unitlet) CreatePod(ctx context.Context, pod *v1.Pod) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	us := units.From(&pod.ObjectMeta, &pod.Spec)
	if err := l.store.CreateUnits(ctx, us); err != nil {
		return err
	}
	for _, u := range us {
		name := u.ID.Name()
		if err := l.state.Link(ctx, l.store.Location(name)); err != nil {
			l.forceUnload(ctx, name)
			return err
		}
		if err := l.state.Enable(ctx, name); err != nil {
			l.forceUnload(ctx, name)
			return err
		}
		if err := l.state.Start(ctx, name); err != nil {
			return err
		}
	}
	return nil
}

func (*Unitlet) UpdatePod(context.Context, *v1.Pod) error { return nil }

func (l *Unitlet) DeletePod(ctx context.Context, pod *v1.Pod) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, u := range units.From(&pod.ObjectMeta, &pod.Spec) {
		name := u.ID.Name()
		if err := l.state.Stop(ctx, name); err != nil {
			continue
		}
		l.forceUnload(ctx, name)
		_ = l.state.Reload(ctx)
	}
	return nil
}

func (l *Unitlet) GetPod(ctx context.Context, namespace, name string) (*v1.Pod, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	//TODO implement me
	panic("implement me")
}

func (l *Unitlet) GetPodStatus(ctx context.Context, namespace, name string) (*v1.PodStatus, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// TODO: It looks up on namespace and name, so how?

	//TODO implement me
	panic("implement me")
}

func (l *Unitlet) GetPods(ctx context.Context) ([]*v1.Pod, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	//TODO implement me
	panic("implement me")
}

func (l *Unitlet) GetContainerLogs(ctx context.Context, namespace, podName, containerName string, opts api.ContainerLogOpts) (io.ReadCloser, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	//TODO implement me
	panic("implement me")
}

func (l *Unitlet) RunInContainer(context.Context, string, string, string, []string, api.AttachIO) error {
	return errs.ErrNotSupported
}

func (l *Unitlet) ConfigureNode(context.Context, *v1.Node) {}

func (l *Unitlet) forceUnload(ctx context.Context, name units.Name) {
	_ = l.state.Disable(ctx, name)
	_ = l.state.ResetFailed(ctx, name)
	_ = l.store.DeleteUnit(ctx, name)
}
