//go:build windows

package sandbox

import (
	"io"
	"os/exec"
	"syscall"
)

func newProcessGroupAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

func killProcessTree(pid int) {
	// Windows: use taskkill to kill the entire process tree
	killCmd := exec.Command("taskkill", "/F", "/T", "/PID", itoa(pid))
	killCmd.Stdout = io.Discard
	killCmd.Stderr = io.Discard
	killCmd.Run()
}
