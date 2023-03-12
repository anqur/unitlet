package units

import (
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func FromPod(om *meta.ObjectMeta, spec *core.PodSpec) (ret []*Unit) {
	for _, c := range spec.Containers {
		var (
			wd   *string
			user *int64
		)
		if c.WorkingDir != "" {
			wd = &c.WorkingDir
		}
		if sc := spec.SecurityContext; sc != nil {
			user = sc.RunAsUser
		}

		ret = append(ret, &Unit{
			ID:     NewID(om.Namespace, om.Name, c.Name),
			Cmd:    append(c.Command, c.Args...),
			PodUID: om.UID,

			Workdir: wd,
			User:    user,
		})
	}
	return
}

func (u *Unit) ToPod(nodeName string, cs []core.Container, status *core.PodStatus) *core.Pod {
	return &core.Pod{
		TypeMeta: meta.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		ObjectMeta: meta.ObjectMeta{
			Name:      u.ID.Pod(),
			Namespace: u.ID.Namespace(),
			UID:       u.PodUID,
		},
		Spec:   core.PodSpec{NodeName: nodeName, Containers: cs},
		Status: *status,
	}
}

func (u *Unit) ToContainer() (ret core.Container) {
	ret.Name = u.ID.Container()
	ret.Command = u.Cmd[:1]
	ret.Args = u.Cmd[1:]
	if wd := u.Workdir; wd != nil {
		ret.WorkingDir = *wd
	}
	if user := u.User; user != nil {
		ret.SecurityContext = &core.SecurityContext{RunAsUser: user}
	}
	return
}

func ToContainerStatus(name string, props Properties, state core.ContainerState) core.ContainerStatus {
	return core.ContainerStatus{
		Name:                 name,
		State:                state,
		LastTerminationState: state,
		Ready:                true,
		RestartCount:         props.RestartCount(),
		ContainerID:          props.ContainerID().String(),
	}
}

func ReduceContainerStatuses(statuses []core.ContainerStatus) core.PodPhase {
	running := 0
	terminated := 0
	isSucceeded := true

	for _, s := range statuses {
		st := s.State
		if st.Waiting != nil {
			return core.PodPending
		}

		if st.Running != nil {
			running++
		}
		if st.Terminated != nil {
			terminated++
			isSucceeded = isSucceeded && st.Terminated.ExitCode == 0
		}
	}

	if running == len(statuses) {
		return core.PodRunning
	}

	if terminated == len(statuses) && isSucceeded {
		return core.PodSucceeded
	}

	return core.PodFailed
}
