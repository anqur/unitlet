package units

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/coreos/go-systemd/v22/unit"
	"k8s.io/apimachinery/pkg/types"

	"github.com/anqur/unitlet/pkg/errs"
)

const (
	ServiceSection = "Service"
	ExecStartKey   = "ExecStart"
	WorkdirKey     = "WorkingDirectory"
	UserKey        = "User"

	K8sSection   = "X-Kubernetes"
	NamespaceKey = "Namespace"
	PodKey       = "Pod"
	PodUIDKey    = "PodUID"
	ContainerKey = "Container"
)

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
				{NamespaceKey, u.ID.Namespace()},
				{PodKey, u.ID.Pod()},
				{PodUIDKey, string(u.PodUID)},
				{ContainerKey, u.ID.Container()},
			},
		},
	}
}

func (u *Unit) UnmarshalUnitSections(ss []*unit.UnitSection) error {
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
				case PodUIDKey:
					u.PodUID = types.UID(e.Value)
				case ContainerKey:
					u.ID.c = e.Value
				}
			}
		}
	}
	return u.checkRequiredFields()
}

func (u *Unit) checkRequiredFields() error {
	if u.ID.ns == "" ||
		u.ID.p == "" ||
		u.ID.c == "" ||
		len(u.Cmd) == 0 ||
		u.PodUID == "" {
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
