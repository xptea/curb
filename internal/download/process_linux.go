package download

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

// yt-dlp starts FFmpeg and a JavaScript runtime. Cancel the entire group so
// a canceled download or closed window leaves no helper processes behind.
func configureProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
}
