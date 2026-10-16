//go:build !windows

package update

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

func processRunning(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}

func configureProcess(command *exec.Cmd) { command.SysProcAttr = &syscall.SysProcAttr{Setsid: true} }
func configureApp(command *exec.Cmd)     { configureProcess(command) }

func cleanupWork(path string) { _ = os.RemoveAll(path) }
