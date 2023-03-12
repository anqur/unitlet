package units

import (
	"fmt"
	"strings"

	"github.com/anqur/unitlet/pkg/errs"
)

const (
	Prefix = "unitlet"
	Suffix = ".service"
	Sep    = "."
)

type ID struct {
	ns, p, c string
}

func NewID(namespace, pod string, container string) ID {
	return ID{namespace, pod, container}
}

func (i *ID) Namespace() string { return i.ns }
func (i *ID) Pod() string       { return i.p }
func (i *ID) Container() string { return i.c }

func (i *ID) String() string { return strings.Join([]string{Prefix, i.ns, i.p, i.c}, Sep) }
func (i *ID) Name() Name     { return Name(i.String() + Suffix) }

func ParseName(name Name) (ID, error) {
	const N = 5
	ss := strings.SplitN(string(name), Sep, N)
	if len(ss) != N || ss[0] != Prefix {
		return ID{}, fmt.Errorf("%w: %s", errs.ErrBadUnitID, name)
	}
	return NewID(ss[1], ss[2], ss[3]), nil
}
