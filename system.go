package pg_info

import (
	"context"
	"os/exec"
	"strings"

	"zgo.at/errors"
)

// System information.
type System struct {
	LoadAvg, Memory, Disk string
}

func (System) Name() string { return "system" }

func (s *System) Data(ctx context.Context) error {
	str := func(b []byte, err error) (string, error) { return string(b), err }

	uptime, err := str(exec.CommandContext(ctx, "uptime").CombinedOutput())
	if err != nil {
		return errors.Wrap(err, "pg_info.System.Data")
	}
	s.LoadAvg = strings.TrimSpace(strings.Join(strings.Split(string(uptime), ",")[2:], ", "))

	s.Memory, err = str(exec.CommandContext(ctx, "free", "-m").CombinedOutput())
	if err != nil {
		return errors.Wrap(err, "pg_info.System.Data")
	}

	// Ignore exit/stderr because:
	// df: /sys/kernel/debug/tracing: Permission denied
	s.Disk, _ = str(exec.CommandContext(ctx, "df", "-hT").Output())

	return nil
}
