package states

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/virtual-kubelet/virtual-kubelet/log"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/anqur/unitlet/pkg/units"
)

const (
	DbusExitCodeKey    = "ExecMainStatus"
	DbusStartedAtKey   = "ExecMainStartTimestamp"
	DbusFinishedAtKey  = "ExecMainExitTimestamp"
	DbusContainerIDKey = "MainPID"

	DbusTerminatedStop   = "stop"
	DbusTerminatedFailed = "failed"
	DbusTerminatedExited = "exited"
	DbusTerminatedDead   = "dead"

	DbusWaitingStart     = "start"
	DbusWaitingCondition = "condition"
	DbusWaitingDead      = DbusTerminatedDead

	DbusRunning            = "running"
	DbusRunningAutoRestart = "auto-restart"
	DbusRunningReload      = "reload"
)

type DbusProperties struct {
	exitCode              int32
	startedAt, finishedAt meta.Time
	containerID           *url.URL
}

func (p *DbusProperties) ExitCode() int32       { return p.exitCode }
func (p *DbusProperties) StartedAt() meta.Time  { return p.startedAt }
func (p *DbusProperties) FinishedAt() meta.Time { return p.finishedAt }
func (p *DbusProperties) ContainerID() *url.URL { return p.containerID }

func (s *DbusState) Properties(ctx context.Context, name units.UnitName) (units.Properties, error) {
	exitCode, err := s.getExitCode(ctx, name)
	if err != nil {
		return nil, err
	}
	startedAt, err := s.getTimeProperty(ctx, name, DbusStartedAtKey)
	if err != nil {
		return nil, err
	}
	finishedAt, err := s.getTimeProperty(ctx, name, DbusFinishedAtKey)
	if err != nil {
		return nil, err
	}
	containerID, err := s.getContainerID(ctx, name)
	if err != nil {
		return nil, err
	}
	return &DbusProperties{
		exitCode:    exitCode,
		startedAt:   startedAt,
		finishedAt:  finishedAt,
		containerID: containerID,
	}, nil
}

func (s *DbusState) getExitCode(ctx context.Context, name units.UnitName) (int32, error) {
	p, err := s.c.GetServicePropertyContext(ctx, string(name), DbusExitCodeKey)
	if err != nil {
		return 0, err
	}
	n, err := strconv.ParseInt(propValue(p), 10, 64)
	if err != nil {
		return -1, nil
	}
	return int32(n), nil
}

func (s *DbusState) getTimeProperty(ctx context.Context, name units.UnitName, key string) (meta.Time, error) {
	p, err := s.c.GetServicePropertyContext(ctx, string(name), key)
	if err != nil {
		return meta.Time{}, err
	}
	t, _ := parseTimestampMilli(propValue(p))
	return meta.NewTime(t), nil
}

func (s *DbusState) getContainerID(ctx context.Context, name units.UnitName) (*url.URL, error) {
	p, err := s.c.GetServicePropertyContext(ctx, string(name), DbusContainerIDKey)
	if err != nil {
		return nil, err
	}
	return &url.URL{Scheme: "pid", Host: propValue(p)}, nil
}

func propValue(p *dbus.Property) string { return fmt.Sprintf("%v", p.Value.Value()) }

func parseTimestampMilli(s string) (time.Time, error) {
	ms, _ := strconv.ParseInt(s, 10, 64)
	return time.UnixMilli(ms), nil
}

func (s *DbusState) toContainerState(props units.Properties, subState string) (ret core.ContainerState) {
	if strings.HasPrefix(subState, DbusTerminatedStop) ||
		subState == DbusTerminatedFailed ||
		subState == DbusTerminatedExited ||
		(subState == DbusTerminatedDead && !props.FinishedAt().IsZero()) {
		reason := string(core.PodSucceeded)
		if props.ExitCode() != 0 {
			reason = string(core.PodFailed)
		}
		ret.Terminated = &core.ContainerStateTerminated{
			ExitCode:    props.ExitCode(),
			Reason:      reason,
			Message:     reason,
			StartedAt:   props.StartedAt(),
			FinishedAt:  props.FinishedAt(),
			ContainerID: props.ContainerID().String(),
		}
		return
	}

	if strings.HasPrefix(subState, DbusWaitingStart) ||
		subState == DbusWaitingCondition ||
		subState == DbusWaitingDead {
		ret.Waiting = &core.ContainerStateWaiting{Reason: subState, Message: subState}
		return
	}

	if subState == DbusRunning ||
		subState == DbusRunningAutoRestart ||
		subState == DbusRunningReload {
		ret.Running = &core.ContainerStateRunning{StartedAt: props.StartedAt()}
		return
	}

	log.L.Warnf("unknown sub-state %q", subState)
	return
}
