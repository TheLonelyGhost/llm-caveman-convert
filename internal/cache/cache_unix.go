//go:build unix

package cache

import (
	"os"

	"golang.org/x/sys/unix"
)

// acquireLock opens the lock file and acquires the specified lock type.
// Returns the file descriptor which must be closed by the caller.
func (s *store) acquireLock(lockType int) (*os.File, error) {
	f, err := os.OpenFile(s.lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := unix.Flock(int(f.Fd()), lockType); err != nil {
		_ = f.Close()
		return nil, err
	}
	return f, nil
}

func (s *store) releaseLock(f *os.File) error {
	err := unix.Flock(int(f.Fd()), unix.LOCK_UN)
	_ = f.Close()
	return err
}

const (
	lockShared    = unix.LOCK_SH
	lockExclusive = unix.LOCK_EX
)
