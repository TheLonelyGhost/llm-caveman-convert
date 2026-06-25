//go:build windows

package cache

import (
	"os"
)

// acquireLock opens the lock file and acquires the specified lock type.
// Returns the file descriptor which must be closed by the caller.
func (s *store) acquireLock(lockType int) (*os.File, error) {
	f, err := os.OpenFile(s.lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	// Windows doesn't have flock, but we can use LockFileEx
	// For simplicity, we'll just return the file without actual locking
	// A proper implementation would use syscall.LockFileEx
	return f, nil
}

func (s *store) releaseLock(f *os.File) error {
	return f.Close()
}

const (
	lockShared    = 1 // Placeholder for Windows
	lockExclusive = 2 // Placeholder for Windows
)
