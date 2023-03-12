package units

import (
	"testing"
)

func TestUnitEncoding(t *testing.T) {
	wd := "/tmp"
	user := int64(42)
	u := &Unit{
		ID:      NewID("a", "b", "c"),
		Cmd:     []string{"echo", "hello"},
		PodUID:  "d",
		Workdir: &wd,
		User:    &user,
	}
	data, err := u.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	u = new(Unit)
	if err := u.Unmarshal(data); err != nil {
		t.Fatal(err)
	}
	if u.ID.String() != "unitlet.a.b.c" ||
		u.Cmd[0] != "echo" ||
		u.Cmd[1] != "hello" ||
		u.PodUID != "d" ||
		*u.Workdir != wd ||
		*u.User != user {
		t.Fatal(u)
	}

	id, err := ParseName(u.ID.Name())
	if err != nil {
		t.Fatal(err)
	}
	if id.Namespace() != "a" ||
		id.Pod() != "b" ||
		id.Container() != "c" {
		t.Fatal(id)
	}
}
