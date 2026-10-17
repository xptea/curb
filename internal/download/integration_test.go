package download

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func integrationTools(t *testing.T) Tools {
	t.Helper()
	if os.Getenv("CURB_INTEGRATION") != "1" {
		t.Skip("set CURB_INTEGRATION=1 for local HTTP and real yt-dlp tests")
	}
	if root := os.Getenv("CURB_INTEGRATION_TOOLS"); root != "" {
		suffix := ""
		if runtime.GOOS == "windows" {
			suffix = ".exe"
		}
		tools := Tools{Downloader: filepath.Join(root, "yt-dlp"+suffix), FFmpeg: filepath.Join(root, "ffmpeg"+suffix), JavaScript: filepath.Join(root, "qjs"+suffix), JSKind: "quickjs"}
		for _, path := range []string{tools.Downloader, tools.FFmpeg, tools.JavaScript} {
			if _, err := os.Stat(path); err != nil {
				t.Fatal(err)
			}
		}
		return tools
	}
	tools, err := LocateTools()
	if err != nil {
		t.Fatal(err)
	}
	return tools
}

func TestRealVideoAndAudioDownload(t *testing.T) {
	tools := integrationTools(t)
	dir := t.TempDir()
	media := filepath.Join(dir, "sample.mp4")
	fixture, err := os.ReadFile("testdata/sample.mp4")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(media, fixture, 0600); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.FileServer(http.Dir(dir)))
	defer server.Close()
	for _, format := range []string{"Video", "Audio"} {
		t.Run(format, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			progress := false
			path, err := Run(ctx, tools, Request{URL: server.URL + "/sample.mp4", Destination: t.TempDir(), Format: format, Quality: "Best available"}, func(u Update) {
				if u.Progress >= 0 {
					progress = true
				}
			})
			if err != nil {
				t.Fatal(err)
			}
			if !progress {
				t.Error("no measurable download progress received")
			}
			ext := ".mp4"
			if format == "Audio" {
				ext = ".mp3"
			}
			if filepath.Ext(path) != ext {
				t.Errorf("expected %s, got %s", ext, path)
			}
			info, err := os.Stat(path)
			if err != nil || info.Size() == 0 {
				t.Fatalf("missing output: %v", err)
			}
		})
	}
	t.Run("HTTP error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, err := Run(ctx, tools, Request{URL: server.URL + "/missing.mp4", Destination: t.TempDir(), Format: "Video"}, func(Update) {})
		if err == nil || !strings.Contains(err.Error(), "404") {
			t.Fatalf("expected actionable HTTP error, got %v", err)
		}
	})
}

func TestRealStreamMerge(t *testing.T) {
	tools := integrationTools(t)
	dir := t.TempDir()
	video, audio, merged := filepath.Join(dir, "video.mp4"), filepath.Join(dir, "audio.m4a"), filepath.Join(dir, "merged.mkv")
	for _, args := range [][]string{
		{"-i", "testdata/sample.mp4", "-map", "0:v:0", "-c", "copy", video},
		{"-i", "testdata/sample.mp4", "-map", "0:a:0", "-c", "copy", audio},
		{"-i", video, "-i", audio, "-map", "0:v:0", "-map", "1:a:0", "-c", "copy", merged},
	} {
		cmd := exec.Command(tools.FFmpeg, append([]string{"-v", "error"}, args...)...)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("stream merge: %v %s", err, output)
		}
	}
	probe := filepath.Join(filepath.Dir(tools.FFmpeg), "ffprobe"+filepath.Ext(tools.FFmpeg))
	output, err := exec.Command(probe, "-v", "error", "-show_entries", "stream=codec_name", "-of", "csv=p=0", merged).CombinedOutput()
	if err != nil || !strings.Contains(string(output), "h264") || !strings.Contains(string(output), "aac") {
		t.Fatalf("merge must retain both original codecs: %v %s", err, output)
	}
}

func TestRealImpersonatedHTTPSDownload(t *testing.T) {
	tools := integrationTools(t)
	server := httptest.NewTLSServer(http.FileServer(http.Dir("testdata")))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	path := filepath.Join(t.TempDir(), "download.mp4")
	// Exercise curl_cffi and its TLS libraries, which ordinary local HTTP
	// downloads do not use. The generated test certificate is self-signed.
	cmd := exec.CommandContext(ctx, tools.Downloader, "--ignore-config", "--no-plugin-dirs", "--no-playlist",
		"--no-check-certificates", "--impersonate", "chrome", "--ffmpeg-location", filepath.Dir(tools.FFmpeg),
		"--output", path, server.URL+"/sample.mp4")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("browser-impersonated HTTPS download: %v %s", err, output)
	}
	want, err := os.ReadFile("testdata/sample.mp4")
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != string(want) {
		t.Fatalf("downloaded media differs from the source fixture: %v", err)
	}
}

func TestCancelRealDownload(t *testing.T) {
	tools := integrationTools(t)
	connected := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case connected <- struct{}{}:
		default:
		}
		<-r.Context().Done()
	}))
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		select {
		case <-connected:
			cancel()
		case <-time.After(10 * time.Second):
			cancel()
		}
	}()
	start := time.Now()
	_, err := Run(ctx, tools, Request{URL: server.URL + "/slow.mp4", Destination: t.TempDir(), Format: "Video"}, func(Update) {})
	if !errors.Is(err, context.Canceled) || time.Since(start) > 12*time.Second {
		t.Fatalf("cancel did not stop promptly: %v", err)
	}
}

func TestRealOutputFormatChoices(t *testing.T) {
	tools := integrationTools(t)
	dir := t.TempDir()
	fixture, err := os.ReadFile("testdata/sample.mp4")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sample.mp4"), fixture, 0600); err != nil {
		t.Fatal(err)
	}
	webm, err := os.ReadFile("testdata/sample.webm")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sample.webm"), webm, 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"-i", "testdata/sample.mp4", "-c", "copy", filepath.Join(dir, "source.mkv")},
		{"-i", "testdata/sample.mp4", "-vn", "-c:a", "libmp3lame", filepath.Join(dir, "source.mp3")},
	} {
		if output, err := exec.Command(tools.FFmpeg, append([]string{"-v", "error"}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("prepare source: %v %s", err, output)
		}
	}
	server := httptest.NewServer(http.FileServer(http.Dir(dir)))
	defer server.Close()
	for _, tc := range []struct{ mode, output, source, extension, codec string }{
		{"Video", "Original", "source.mkv", ".mkv", "h264"},
		{"Video", "MP4", "source.mkv", ".mp4", "h264"},
		{"Video", "MP4", "sample.webm", ".mp4", "h264"},
		{"Video", "Original", "sample.webm", ".webm", "vp9"},
		{"Video", "MKV", "sample.mp4", ".mkv", "h264"},
		{"Audio", "Original", "source.mp3", ".mp3", "mp3"},
		{"Audio", "Original", "sample.webm", ".opus", "opus"},
		{"Audio", "MP3", "source.mkv", ".mp3", "mp3"},
		{"Audio", "M4A", "source.mp3", ".m4a", "aac"},
		{"Audio", "WAV", "source.mkv", ".wav", "pcm_s16le"},
		{"Audio", "FLAC", "source.mkv", ".flac", "flac"},
	} {
		t.Run(tc.mode+"/"+tc.output, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			path, err := Run(ctx, tools, Request{URL: server.URL + "/" + tc.source, Destination: t.TempDir(), Format: tc.mode, VideoOutputFormat: tc.output, AudioOutputFormat: tc.output}, func(Update) {})
			if err != nil {
				t.Fatal(err)
			}
			if filepath.Ext(path) != tc.extension {
				t.Fatalf("expected %s, got %s", tc.extension, path)
			}
			probe := filepath.Join(filepath.Dir(tools.FFmpeg), "ffprobe"+filepath.Ext(tools.FFmpeg))
			output, err := exec.Command(probe, "-v", "error", "-show_entries", "stream=codec_name", "-of", "csv=p=0", path).CombinedOutput()
			if err != nil || !strings.Contains(string(output), tc.codec) {
				t.Fatalf("wrong final codec: %v %s", err, output)
			}
			if tc.mode == "Audio" && strings.Contains(string(output), "h264") {
				t.Fatal("audio download still contains video")
			}
			if tc.mode == "Video" && tc.output == "MP4" && !strings.Contains(string(output), "aac") {
				t.Fatal("converted MP4 is missing AAC audio")
			}
			if tc.output == "Original" && (tc.mode == "Video" || tc.source == "source.mp3") {
				want, _ := os.ReadFile(filepath.Join(dir, tc.source))
				got, _ := os.ReadFile(path)
				if string(want) != string(got) {
					t.Fatal("original media was re-encoded or repackaged")
				}
			}
		})
	}
	t.Run("Preserve existing video when extracting audio", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		destination := t.TempDir()
		request := Request{URL: server.URL + "/sample.mp4", Destination: destination, Format: "Video", VideoOutputFormat: "Original"}
		path, err := Run(ctx, tools, request, func(Update) {})
		if err != nil {
			t.Fatal(err)
		}
		request.Format, request.AudioOutputFormat = "Audio", "MP3"
		if _, err := Run(ctx, tools, request, func(Update) {}); err != nil {
			t.Fatal(err)
		}
		preserved, err := os.ReadFile(path)
		if err != nil || string(preserved) != string(fixture) {
			t.Fatalf("audio conversion changed or removed an existing video: %v", err)
		}
	})
}
