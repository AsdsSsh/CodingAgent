//go:build !windows

package sandbox

import (
	"syscall"
)

func newProcessGroupAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}

func killProcessTree(pid int) {
	// Kill the entire process group (negative PID)
	syscall.Kill(-pid, syscall.SIGKILL)
}
