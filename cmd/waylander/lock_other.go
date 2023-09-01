//go:build !(aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos)

package main

func GetLock() bool {
	return true
}

func ReleaseLock() {}
