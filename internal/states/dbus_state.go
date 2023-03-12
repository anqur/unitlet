package states

import (
	"context"
	"fmt"
	"strings"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/coreos/go-systemd/v22/util"
	"github.com/virtual-kubelet/virtual-kubelet/errdefs"
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

func (s *DbusState) List(ctx context.Context) (map[units.Name]*core.PodStatus, error) {
	infos, err := s.listUnits(ctx)
	if err != nil {
		return nil, err
	}
	ret := make(map[units.Name]*core.PodStatus)
	for name, info := range infos {
		phase := units.ReduceContainerStatuses(info.statuses)
		startedAt := info.props.StartedAt()
		ret[name] = &core.PodStatus{
			Phase: phase,
			Conditions: []core.PodCondition{
				{Type: core.PodReady, Status: core.ConditionTrue, LastTransitionTime: startedAt},
				{Type: core.PodInitialized, Status: core.ConditionTrue, LastTransitionTime: startedAt},
				{Type: core.PodScheduled, Status: core.ConditionTrue, LastTransitionTime: startedAt},
			},
			Message:   string(phase),
			StartTime: &startedAt,
		}
	}
	return ret, nil
}

type unitInfo struct {
	props    units.Properties
	statuses []core.ContainerStatus
}

func (s *DbusState) listUnits(ctx context.Context) (map[units.Name]*unitInfo, error) {
	us, err := s.c.ListUnitsContext(ctx)
	if err != nil {
		return nil, err
	}

	unitInfos := make(map[units.Name]*unitInfo)
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

		state := fromSubState(u.SubState, props)
		status := units.ToContainerStatus(id.Container(), props, state)

		info, ok := unitInfos[name]
		if !ok {
			info = &unitInfo{props: props}
			unitInfos[name] = info
		}
		info.statuses = append(info.statuses, status)
	}

	return unitInfos, nil
}

func (s *DbusState) Get(ctx context.Context, name units.Name) (*core.PodStatus, error) {
	m, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	if status, ok := m[name]; ok {
		return status, nil
	}
	return nil, errdefs.NotFound(string(name))
}
