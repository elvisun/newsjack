//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/windows"
)

func lockFileWithTimeout(file *os.File, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		overlapped := new(windows.Overlapped)
		err := windows.LockFileEx(
			windows.Handle(file.Fd()),
			windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY,
			0,
			1,
			0,
			overlapped,
		)
		if err == nil {
			return nil
		}
		if !errors.Is(err, windows.ERROR_LOCK_VIOLATION) && !errors.Is(err, windows.ERROR_IO_PENDING) {
			return err
		}
		if !time.Now().Before(deadline) {
			return fmt.Errorf("timed out after %s", timeout)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func unlockFile(file *os.File) error {
	return windows.UnlockFileEx(windows.Handle(file.Fd()), 0, 1, 0, new(windows.Overlapped))
}
