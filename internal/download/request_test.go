package download

import (
	"strings"
	"testing"
)

func TestURLValidation(t *testing.T) {
	for _, raw := range []string{"", "a video", "file:///etc/passwd", "https://", "ftp://example.com/a", "--exec rm", "https:///missing-host"} {
		if _, err := ValidateURL(raw); err == nil {
			t.Errorf("accepted %q", raw)
		}
	}
	for _, raw := range []string{"https://example.com/watch?v=1&list=x", " http://localhost:8080/video.mp4 \n"} {
		if link, err := ValidateURL(raw); err != nil || link != strings.TrimSpace(raw) {
			t.Errorf("rejected %q: %v", raw, err)
		}
	}
}

func TestDownloadArguments(t *testing.T) {
	tools := Tools{Downloader: "/tools/yt-dlp", FFmpeg: "/tools/ffmpeg", JavaScript: "/tools/qjs", JSKind: "quickjs"}
	r := Request{URL: "https://example.com/video?x=$(touch bad)", Destination: "/tmp/folder with spaces", Format: "Video", Quality: "720p"}
	args := downloadArgs(tools, r)
	joined := strings.Join(args, "\n")
	for _, want := range []string{"--ignore-config", "--no-playlist", "--no-overwrites", "bv*[height<=720]+ba/b[height<=720]", "--no-js-runtimes\n--js-runtimes\nquickjs:/tools/qjs", r.Destination} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %s", want)
		}
	}
	if args[len(args)-2] != "--" || args[len(args)-1] != r.URL {
		t.Fatal("URL must be one argument after option delimiter")
	}
	r.Format = "Audio"
	joined = strings.Join(downloadArgs(tools, r), "\n")
	if !strings.Contains(joined, "--extract-audio\n--audio-format\nmp3") || strings.Contains(joined, "--merge-output-format") {
		t.Fatal("wrong audio arguments")
	}
	r.AudioQuality = "192 kbps"
	if !strings.Contains(strings.Join(downloadArgs(tools, r), "\n"), "--audio-quality\n192K") {
		t.Fatal("selected audio quality not applied")
	}
}
