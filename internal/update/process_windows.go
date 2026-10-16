package update

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unicode/utf16"
)

var kernel = syscall.NewLazyDLL("kernel32.dll")
var waitProcess = kernel.NewProc("WaitForSingleObject")

func processRunning(pid int) bool {
	const synchronize = 0x00100000
	handle, err := syscall.OpenProcess(synchronize, false, uint32(pid))
	if err != nil {
		return errors.Is(err, syscall.ERROR_ACCESS_DENIED)
	}
	defer syscall.CloseHandle(handle)
	status, _, _ := waitProcess.Call(uintptr(handle), 0)
	return status != 0 // Only WAIT_OBJECT_0 proves that the process exited.
}

func configureProcess(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}

func configureApp(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
}

func cleanupWork(path string) {
	// Windows cannot delete the running helper. A short-lived PowerShell
	// process waits for it to exit, then removes only this private work folder.
	literal := "'" + strings.ReplaceAll(path, "'", "''") + "'"
	script := "Wait-Process -Id " + strconv.Itoa(os.Getpid()) + " -ErrorAction SilentlyContinue; for ($i=0; $i -lt 20; $i++) { Remove-Item -LiteralPath " + literal + " -Recurse -Force -ErrorAction SilentlyContinue; if (!(Test-Path -LiteralPath " + literal + ")) { break }; Start-Sleep -Milliseconds 250 }"
	units := utf16.Encode([]rune(script))
	data := make([]byte, len(units)*2)
	for i, unit := range units {
		binary.LittleEndian.PutUint16(data[i*2:], unit)
	}
	shell := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	command := exec.Command(shell, "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-EncodedCommand", base64.StdEncoding.EncodeToString(data))
	configureProcess(command)
	command.Dir, command.Env = os.TempDir(), cleanEnvironment()
	if command.Start() == nil {
		_ = command.Process.Release()
	}
}
