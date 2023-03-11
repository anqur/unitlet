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

func (s *DbusState) Enable(ctx context.Context, name units.UnitName) error {
	ok, _, err := s.c.EnableUnitFilesContext(ctx, []string{string(name)}, true, true)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: %s", errs.ErrDbusEnable, name)
	}
	return nil
}

func (s *DbusState) Disable(ctx context.Context, name units.UnitName) error {
	_, err := s.c.DisableUnitFilesContext(ctx, []string{string(name)}, true)
	return err
}

func (s *DbusState) Start(ctx context.Context, name units.UnitName) error {
	_, err := s.c.StartUnitContext(ctx, string(name), DbusModeReplace, nil)
	return err
}

func (s *DbusState) Stop(ctx context.Context, name units.UnitName) error {
	_, err := s.c.StopUnitContext(ctx, string(name), DbusModeReplace, nil)
	return err
}

func (s *DbusState) Reload(ctx context.Context) error { return s.c.ReloadContext(ctx) }

func (s *DbusState) ResetFailed(ctx context.Context, name units.UnitName) error {
	return s.c.ResetFailedUnitContext(ctx, string(name))
}

func (s *DbusState) List(ctx context.Context) (map[units.UnitName]*core.Pod, error) {
	us, err := s.c.ListUnitsContext(ctx)
	if err != nil {
		return nil, err
	}
	for _, u := range us {
		if !strings.HasPrefix(u.Name, units.UnitPrefix) {
			continue
		}
		if !strings.HasSuffix(u.Name, units.UnitSuffix) {
			continue
		}
		// TODO
	}
	return nil, nil
}

func (s *DbusState) Get(ctx context.Context, name units.UnitName) (*core.Pod, error) {
	m, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	if info, ok := m[name]; ok {
		return info, nil
	}
	return nil, errdefs.NotFound(string(name))
}
