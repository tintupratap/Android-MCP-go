//go:build windows

package scrcpy

import (
	"fmt"
	"os"
	"syscall"
)

var (
	modkernel32  = syscall.NewLazyDLL("kernel32.dll")
	procLockFile = modkernel32.NewProc("LockFile")
)

func lockFile(file *os.File) error {
	r1, _, err := procLockFile.Call(file.Fd(), 0, 0, 1, 0)
	if r1 == 0 {
		if err != nil && err.Error() != "The operation completed successfully." {
			return err
		}
		return fmt.Errorf("failed to lock file")
	}
	return nil
}
