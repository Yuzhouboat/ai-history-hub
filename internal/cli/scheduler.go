package cli

import (
	"fmt"
	"time"

	"claude-backup/internal/system"
)

// scheduledHour and scheduledMinute are the fixed daily time, in the host's
// local time zone, at which claude-backup's generated schedule (launchd on
// darwin, systemd on linux) invokes sync — both StartCalendarInterval and
// unqualified OnCalendar values are interpreted in local time by their
// respective schedulers, so nextDailyRun matches that rather than fixing UTC.
const (
	scheduledHour   = 3
	scheduledMinute = 0
)

// scheduleStatus is what a platform scheduler reports back to the status
// subcommand. The zero value means "not scheduled".
type scheduleStatus struct {
	Active  bool
	NextRun time.Time
}

// scheduler wires claude-backup's daily sync into (and out of) the host
// platform's own scheduler: launchd on darwin, systemd on linux.
type scheduler interface {
	// install generates (or updates in place) the platform schedule and
	// activates it immediately.
	install(sys system.System, homeDir string) error

	// uninstall deactivates and removes the platform schedule. It is a
	// no-op if nothing was ever installed. It never touches S3 data or the
	// backup config.
	uninstall(sys system.System, homeDir string) error

	// status reports whether the schedule is active and when it next runs.
	status(sys system.System, homeDir string) (scheduleStatus, error)
}

// schedulerFor selects the scheduler implementation for goos (normally
// runtime.GOOS). darwinScheduler and linuxScheduler are exercised directly
// in tests so each platform's behavior is covered through the fake System
// regardless of which platform the test suite itself runs on.
func schedulerFor(goos string) (scheduler, error) {
	switch goos {
	case "darwin":
		return darwinScheduler{}, nil
	case "linux":
		return linuxScheduler{}, nil
	default:
		return nil, fmt.Errorf("automatic scheduling is not supported on %q", goos)
	}
}

// nextDailyRun returns the next occurrence of the daily schedule time at or
// after now, in now's own location — matching the local-time interpretation
// launchd/systemd apply to the generated schedule.
func nextDailyRun(now time.Time) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), scheduledHour, scheduledMinute, 0, 0, now.Location())
	if !next.After(now) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}
