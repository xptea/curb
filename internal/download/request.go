package download

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
)

type Request struct{ URL, Destination, Format, Quality, AudioQuality string }

func ValidateURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	u, err := url.ParseRequestURI(raw)
	if err != nil || u.Hostname() == "" || (u.Scheme != "https" && u.Scheme != "http") {
		return "", errors.New("Paste a valid video or audio link starting with https:// or http://.")
	}
	return raw, nil
}

func downloadArgs(t Tools, r Request) []string {
	args := []string{"--ignore-config", "--no-playlist", "--no-colors", "--newline", "--progress", "--progress-delta", "0.3",
		"--no-simulate", "--no-overwrites", "--socket-timeout", "20", "--retries", "3", "--windows-filenames",
		"--ffmpeg-location", filepath.Dir(t.FFmpeg),
		"--paths", r.Destination, "--output", "%(title).160B.%(ext)s",
		"--progress-template", "download:curb-progress:%(progress)j",
		"--print", "before_dl:curb-title:%(title)j", "--print", "after_move:curb-file:%(filepath)j"}
	if t.JavaScript != "" {
		args = append(args, "--no-js-runtimes", "--js-runtimes", t.JSKind+":"+t.JavaScript)
	}
	if r.Format == "Audio" {
		quality := "0"
		switch r.AudioQuality {
		case "320 kbps", "256 kbps", "192 kbps", "128 kbps":
			quality = strings.TrimSuffix(r.AudioQuality, " kbps") + "K"
		}
		args = append(args, "--extract-audio", "--audio-format", "mp3", "--audio-quality", quality)
	} else {
		selector := "bv*+ba/b"
		if height, err := strconv.Atoi(strings.TrimSuffix(r.Quality, "p")); err == nil {
			selector = fmt.Sprintf("bv*[height<=%d]+ba/b[height<=%d]", height, height)
		}
		args = append(args, "--format", selector, "--merge-output-format", "mkv")
	}
	return append(args, "--", r.URL)
}
