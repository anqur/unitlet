package providers

import (
	"context"
	"io"
	"sync"

	"github.com/virtual-kubelet/node-cli/provider"
	"github.com/virtual-kubelet/virtual-kubelet/errdefs"
	"github.com/virtual-kubelet/virtual-kubelet/node/api"
	core "k8s.io/api/core/v1"

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

func (l *Unitlet) CreatePod(ctx context.Context, pod *core.Pod) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	us := units.FromPod(&pod.ObjectMeta, &pod.Spec)
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

func (*Unitlet) UpdatePod(context.Context, *core.Pod) error { return nil }

func (l *Unitlet) DeletePod(ctx context.Context, pod *core.Pod) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, u := range units.FromPod(&pod.ObjectMeta, &pod.Spec) {
		name := u.ID.Name()
		if err := l.state.Stop(ctx, name); err != nil {
			continue
		}
		l.forceUnload(ctx, name)
		_ = l.state.Reload(ctx)
	}
	return nil
}

func (l *Unitlet) GetPod(ctx context.Context, namespace, name string) (*core.Pod, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	view, err := l.getView(ctx, namespace, name)
	if err != nil {
		return nil, err
	}
	cs, err := l.getContainers(ctx, view.Names)
	if err != nil {
		return nil, err
	}
	lead, err := l.store.GetUnit(ctx, view.Lead)
	if err != nil {
		return nil, err
	}
	return lead.ToPod(l.cfg.NodeName, cs, view.Status), nil
}

func (l *Unitlet) GetPodStatus(ctx context.Context, namespace, name string) (*core.PodStatus, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	view, err := l.getView(ctx, namespace, name)
	if err != nil {
		return nil, err
	}
	return view.Status, nil
}

func (l *Unitlet) GetPods(ctx context.Context) (ret []*core.Pod, err error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	views, err := l.state.Views(ctx)
	if err != nil {
		return nil, err
	}
	for _, pods := range views {
		for _, view := range pods {
			cs, err := l.getContainers(ctx, view.Names)
			if err != nil {
				return nil, err
			}
			lead, err := l.store.GetUnit(ctx, view.Lead)
			if err != nil {
				return nil, err
			}
			ret = append(ret, lead.ToPod(l.cfg.NodeName, cs, view.Status))
		}
	}
	return
}

func (l *Unitlet) getView(ctx context.Context, namespace, name string) (*units.View, error) {
	views, err := l.state.Views(ctx)
	if err != nil {
		return nil, err
	}
	pods, ok := views[namespace]
	if !ok {
		return nil, errdefs.NotFound(namespace)
	}
	view, ok := pods[name]
	if ok {
		return nil, errdefs.NotFound(name)
	}
	return view, nil
}

func (l *Unitlet) getContainers(ctx context.Context, names []units.Name) (ret []core.Container, err error) {
	for _, name := range names {
		var u *units.Unit
		if u, err = l.store.GetUnit(ctx, name); err != nil {
			return
		}
		ret = append(ret, u.ToContainer())
	}
	return
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

func (l *Unitlet) ConfigureNode(context.Context, *core.Node) {}

func (l *Unitlet) forceUnload(ctx context.Context, name units.Name) {
	_ = l.state.Disable(ctx, name)
	_ = l.state.ResetFailed(ctx, name)
	_ = l.store.DeleteUnit(ctx, name)
}
