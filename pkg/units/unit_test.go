package units

import (
	"testing"
)

func TestUnitEncoding(t *testing.T) {
	wd := "/tmp"
	user := int64(42)
	u := &Unit{
		ID:      NewUnitID("a", "b", "c"),
		Cmd:     []string{"echo", "hello"},
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
	if u.ID.String() != "a.b.c" ||
		u.Cmd[0] != "echo" ||
		u.Cmd[1] != "hello" ||
		*u.Workdir != wd ||
		*u.User != user {
		t.Fatal(u)
	}
}
