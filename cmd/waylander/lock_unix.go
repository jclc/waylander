//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos

package main

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

const fileLockPath = "/tmp/waylander.lock"

var lockFile *os.File

// GetLock acquires a filesystem lock in order to prevent waylander from
// accidentally running conflicting desktop session operations simultaneously.
func GetLock() error {
	f, _ := os.Create(fileLockPath)
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		return errors.New(
			"filesystem lock is taken; if waylander is not currently running, delete " + fileLockPath)
	}

	lockFile = f
	return nil
}

func ReleaseLock() {
	_ = unix.Flock(int(lockFile.Fd()), unix.LOCK_UN)
	_ = os.Remove(fileLockPath)
}
