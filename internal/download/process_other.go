//go:build !linux && !darwin && !windows

package download

import "os/exec"

func configureProcess(cmd *exec.Cmd) {}
