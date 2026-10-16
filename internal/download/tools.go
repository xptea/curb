package download

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type Tools struct{ Downloader, FFmpeg, JavaScript, JSKind string }

func LocateTools() (Tools, error) {
	exe, err := os.Executable()
	if err != nil {
		return Tools{}, err
	}
	if resolved, e := filepath.EvalSymlinks(exe); e == nil {
		exe = resolved
	}
	root := filepath.Dir(exe)
	find := func(name string) string {
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		for _, dir := range []string{filepath.Join(root, "tools"), filepath.Join(root, "..", "tools"), filepath.Join(root, "..", "Resources", "tools"), developmentToolsDirectory()} {
			if dir == "" {
				continue
			}
			p, e := filepath.Abs(filepath.Join(dir, name))
			if e == nil {
				if info, e := os.Stat(p); e == nil && info.Mode().IsRegular() && (runtime.GOOS == "windows" || info.Mode().Perm()&0111 != 0) {
					return p
				}
			}
		}
		p, _ := exec.LookPath(name)
		return p
	}
	t := Tools{Downloader: find("yt-dlp"), FFmpeg: find("ffmpeg")}
	for _, runtime := range []struct{ kind, executable string }{{"quickjs", "qjs"}, {"deno", "deno"}, {"node", "node"}} {
		if p := find(runtime.executable); p != "" {
			t.JavaScript, t.JSKind = p, runtime.kind
			break
		}
	}
	if t.Downloader == "" {
		return t, errors.New("Download engine is missing. Run python3 scripts/bundle.py to prepare the bundled tools.")
	}
	if t.FFmpeg == "" {
		return t, errors.New("FFmpeg is missing. Run python3 scripts/bundle.py to prepare video and audio support.")
	}
	return t, nil
}

// Tests run inside package directories and go run uses a temporary executable.
// Find the module root for development without relying on a fixed working directory.
func developmentToolsDirectory() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if info, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && info.Mode().IsRegular() {
			return filepath.Join(dir, "tools")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
