package states

import (
	"context"
	"fmt"
	"strings"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/coreos/go-systemd/v22/util"
	core "k8s.io/api/core/v1"

	"github.com/anqur/unitlet/pkg/errs"
	"github.com/anqur/unitlet/pkg/units"
)

const DbusModeReplace = "replace"

type DbusState struct{ c *dbus.Conn }

func NewDbusState() (units.State, error) {
	if !util.IsRunningSystemd() {
		return nil, errs.ErrSystemdNotRunning
	}
	c, err := dbus.NewWithContext(context.Background())
	if err != nil {
		return nil, err
	}
	return &DbusState{c}, nil
}

func (s *DbusState) Link(ctx context.Context, loc units.Location) error {
	_, err := s.c.LinkUnitFilesContext(ctx, []string{string(loc)}, true, true)
	return err
}

func (s *DbusState) Enable(ctx context.Context, name units.Name) error {
	ok, _, err := s.c.EnableUnitFilesContext(ctx, []string{string(name)}, true, true)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: %s", errs.ErrDbusEnable, name)
	}
	return nil
}

func (s *DbusState) Disable(ctx context.Context, name units.Name) error {
	_, err := s.c.DisableUnitFilesContext(ctx, []string{string(name)}, true)
	return err
}

func (s *DbusState) Start(ctx context.Context, name units.Name) error {
	_, err := s.c.StartUnitContext(ctx, string(name), DbusModeReplace, nil)
	return err
}

func (s *DbusState) Stop(ctx context.Context, name units.Name) error {
	_, err := s.c.StopUnitContext(ctx, string(name), DbusModeReplace, nil)
	return err
}

func (s *DbusState) Reload(ctx context.Context) error { return s.c.ReloadContext(ctx) }

func (s *DbusState) ResetFailed(ctx context.Context, name units.Name) error {
	return s.c.ResetFailedUnitContext(ctx, string(name))
}

func (s *DbusState) Views(ctx context.Context) (units.Views, error) {
	infos, err := s.listUnits(ctx)
	if err != nil {
		return nil, err
	}

	namespaces := make(units.Views)
	for _, info := range infos {
		ns := info.id.Namespace()
		pod := info.id.Pod()
		name := info.id.Name()
		var (
			ok   bool
			pods map[string]*units.View
			view *units.View
		)
		if pods, ok = namespaces[ns]; !ok {
			pods = make(map[string]*units.View)
			namespaces[ns] = pods
		}
		if view, ok = pods[pod]; !ok {
			startedAt := info.props.StartedAt()
			view = &units.View{
				Lead: name,
				Status: &core.PodStatus{
					Conditions: []core.PodCondition{
						{Type: core.PodReady, Status: core.ConditionTrue, LastTransitionTime: startedAt},
						{Type: core.PodInitialized, Status: core.ConditionTrue, LastTransitionTime: startedAt},
						{Type: core.PodScheduled, Status: core.ConditionTrue, LastTransitionTime: startedAt},
					},
				},
			}
			pods[pod] = view
		}
		view.Names = append(view.Names, name)
		view.Status.ContainerStatuses = append(view.Status.ContainerStatuses, info.status)
	}

	for _, pods := range namespaces {
		for _, view := range pods {
			phase := units.ReduceContainerStatuses(view.Status.ContainerStatuses)
			view.Status.Phase = phase
			view.Status.Message = string(phase)
		}
	}

	return namespaces, nil
}

type unitStatus struct {
	id     units.ID
	props  units.Properties
	status core.ContainerStatus
}

func (s *DbusState) listUnits(ctx context.Context) ([]*unitStatus, error) {
	us, err := s.c.ListUnitsContext(ctx)
	if err != nil {
		return nil, err
	}

	var ret []*unitStatus
	for _, u := range us {
		if !strings.HasPrefix(u.Name, units.Prefix) {
			continue
		}
		if !strings.HasSuffix(u.Name, units.Suffix) {
			continue
		}

		name := units.Name(u.Name)
		id, err := units.ParseName(name)
		if err != nil {
			return nil, err
		}
		props, err := s.Properties(ctx, name)
		if err != nil {
			return nil, err
		}

		ret = append(ret, &unitStatus{
			id,
			props,
			units.ToContainerStatus(
				id.Container(),
				props,
				toContainerState(u.SubState, props),
			),
		})
	}
	return ret, nil
}
