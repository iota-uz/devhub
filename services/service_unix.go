//go:build !windows

package services

import (
	"syscall"
)

func setSysProcAttr(attr *syscall.SysProcAttr) {
	attr.Setpgid = true
}

func killProcessGroup(pid int32) error {
	return syscall.Kill(-int(pid), syscall.SIGKILL)
}
