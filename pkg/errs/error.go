package errs

import (
	"errors"
	"fmt"
)

var (
	Err = errors.New("unitlet error")

	ErrNotSupported = wrap("not supported")

	ErrBadUnitFile = wrap("not a Pod-compatible unit file")
	ErrBadUnitID   = wrap("invalid unit ID")

	ErrUnitFileExists  = wrap("unit file already exists")
	ErrMarshalUnitFile = wrap("unit file marshal error")
	ErrWriteUnitFile   = wrap("unit file write error")

	ErrSystemdNotRunning = wrap("systemd not running")
	ErrDbusEnable        = wrap("dbus enable error")
)

func wrap(msg string) error { return fmt.Errorf("%w: %s", Err, msg) }
