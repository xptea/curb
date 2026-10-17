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

func TestOutputFormatArguments(t *testing.T) {
	tools := Tools{FFmpeg: "/tools/ffmpeg"}
	request := Request{URL: "https://example.com/media", Destination: "/downloads", Format: "Video"}
	for _, output := range []string{"Original", "MP4", "MKV"} {
		request.VideoOutputFormat = output
		args := strings.Join(downloadArgs(tools, request), "\n")
		switch output {
		case "Original":
			if strings.Contains(args, "--recode-video") || strings.Contains(args, "--remux-video") || strings.Contains(args, "--merge-output-format") {
				t.Fatal("original video must not force a container or conversion")
			}
		case "MP4":
			for _, want := range []string{"--recode-video\nmp4", "VideoConvertor+ffmpeg_o:", "-c:v libx264", "-c:a aac", "-pix_fmt yuv420p"} {
				if !strings.Contains(args, want) {
					t.Fatalf("MP4 conversion missing %q", want)
				}
			}
		case "MKV":
			if !strings.Contains(args, "--remux-video\nmkv") || strings.Contains(args, "--recode-video") {
				t.Fatal("MKV should repackage without re-encoding")
			}
		}
	}
	request.Format = "Audio"
	request.AudioQuality = "192 kbps"
	for _, output := range []string{"Original", "MP3", "M4A", "WAV", "FLAC"} {
		request.AudioOutputFormat = output
		args := strings.Join(downloadArgs(tools, request), "\n")
		format := strings.ToLower(output)
		if output == "Original" {
			format = "best"
		}
		if !strings.Contains(args, "--audio-format\n"+format) || !strings.Contains(args, "--format\nbestaudio/best") {
			t.Fatalf("wrong extraction for %s: %s", output, args)
		}
		if strings.Contains(args, "--audio-quality") != (output == "MP3" || output == "M4A") {
			t.Fatalf("bitrate applied incorrectly to %s", output)
		}
	}
}
