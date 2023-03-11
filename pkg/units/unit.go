package units

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/coreos/go-systemd/v22/unit"
	core "k8s.io/api/core/v1"

	"github.com/anqur/unitlet/pkg/errs"
)

const (
	Prefix = "unitlet"
	Suffix = ".service"
)

type ID struct {
	ns, p, c string
}

func NewID(namespace, pod, container string) *ID {
	return &ID{namespace, pod, container}
}

func (i *ID) Namespace() string { return i.ns }
func (i *ID) Pod() string       { return i.p }
func (i *ID) Container() string { return i.c }
func (i *ID) String() string    { return strings.Join([]string{Prefix, i.ns, i.p, i.c}, ".") }
func (i *ID) Name() Name        { return Name(i.String() + Suffix) }

type Name string

const (
	ServiceSection = "Service"
	ExecStartKey   = "ExecStart"
	WorkdirKey     = "WorkingDirectory"
	UserKey        = "User"

	K8sSection   = "X-Kubernetes"
	NamespaceKey = "Namespace"
	PodKey       = "Pod"
	ContainerKey = "Container"
)

type Unit struct {
	ID  *ID
	Cmd []string

	Workdir *string
	User    *int64
}

func FromPod(p *core.Pod) (ret []*Unit) {
	for _, c := range p.Spec.Containers {
		var (
			wd   *string
			user *int64
		)
		if c.WorkingDir != "" {
			wd = &c.WorkingDir
		}
		if sc := p.Spec.SecurityContext; sc != nil {
			user = sc.RunAsUser
		}

		ret = append(ret, &Unit{
			ID:  NewID(p.Namespace, p.Name, c.Name),
			Cmd: append(c.Command, c.Args...),

			Workdir: wd,
			User:    user,
		})
	}
	return
}

func (u *Unit) MarshalUnitSections() []*unit.UnitSection {
	serviceEntries := []*unit.UnitEntry{
		{"Type", "simple"},
		{ExecStartKey, strings.Join(u.Cmd, " ")},
	}
	if wd := u.Workdir; wd != nil {
		serviceEntries = append(serviceEntries, &unit.UnitEntry{
			Name:  WorkdirKey,
			Value: *wd,
		})
	}
	if user := u.User; user != nil {
		serviceEntries = append(serviceEntries, &unit.UnitEntry{
			Name:  UserKey,
			Value: strconv.FormatInt(*user, 10),
		})
	}

	return []*unit.UnitSection{
		{
			Section: "Unit",
			Entries: []*unit.UnitEntry{
				{"Description", u.ID.String()},
				{"After", "network-online.target"},
			},
		},
		{
			Section: ServiceSection,
			Entries: serviceEntries,
		},
		{
			Section: "Install",
			Entries: []*unit.UnitEntry{
				{"WantedBy", "multi-user.target"},
			},
		},
		{
			Section: K8sSection,
			Entries: []*unit.UnitEntry{
				{NamespaceKey, u.ID.ns},
				{PodKey, string(u.ID.p)},
				{ContainerKey, u.ID.c},
			},
		},
	}
}

func (u *Unit) UnmarshalUnitSections(ss []*unit.UnitSection) error {
	u.ID = new(ID)
	for _, s := range ss {
		for _, e := range s.Entries {
			switch s.Section {
			case ServiceSection:
				switch e.Name {
				case ExecStartKey:
					u.Cmd = strings.Split(e.Value, " ")

				case WorkdirKey:
					u.Workdir = &e.Value
				case UserKey:
					user, err := strconv.ParseInt(e.Value, 10, 64)
					if err != nil {
						return fmt.Errorf("%w: u=%+v, err=%v", errs.ErrBadUnitFile, u, err)
					}
					u.User = &user
				}
			case K8sSection:
				switch e.Name {
				case NamespaceKey:
					u.ID.ns = e.Value
				case PodKey:
					u.ID.p = e.Value
				case ContainerKey:
					u.ID.c = e.Value
				}
			}
		}
	}
	return u.checkRequiredFields()
}

func (u *Unit) checkRequiredFields() error {
	if len(u.Cmd) == 0 || u.ID.ns == "" || u.ID.p == "" || u.ID.c == "" {
		return fmt.Errorf("%w: u=%+v", errs.ErrBadUnitFile, u)
	}
	return nil
}

func (u *Unit) Marshal() ([]byte, error) {
	return io.ReadAll(unit.SerializeSections(u.MarshalUnitSections()))
}

func (u *Unit) Unmarshal(data []byte) error {
	ss, err := unit.DeserializeSections(bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("%w: %v", errs.ErrBadUnitFile, err)
	}
	return u.UnmarshalUnitSections(ss)
}
