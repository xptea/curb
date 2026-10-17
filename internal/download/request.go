package download

import (
	"curb/internal/media"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
)

type Request struct{ URL, Destination, Format, Quality, AudioQuality, VideoOutputFormat, AudioOutputFormat string }

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
		output := media.AudioOutput(r.AudioOutputFormat)
		format := strings.ToLower(output)
		if output == media.Original {
			format = "best"
		}
		args = append(args, "--format", "bestaudio/best", "--extract-audio", "--audio-format", format)
		if media.AudioHasBitrate(output) {
			args = append(args, "--audio-quality", quality)
		}
	} else {
		selector := "bv*+ba/b"
		if height, err := strconv.Atoi(strings.TrimSuffix(r.Quality, "p")); err == nil {
			selector = fmt.Sprintf("bv*[height<=%d]+ba/b[height<=%d]", height, height)
		}
		args = append(args, "--format", selector)
		switch media.VideoOutput(r.VideoOutputFormat) {
		case "MP4":
			args = append(args, "--recode-video", "mp4", "--postprocessor-args",
				"VideoConvertor+ffmpeg_o:-c:v libx264 -preset fast -crf 20 -pix_fmt yuv420p -vf scale=ceil(iw/2)*2:ceil(ih/2)*2 -c:a aac -movflags +faststart")
		case "MKV":
			args = append(args, "--merge-output-format", "mkv", "--remux-video", "mkv")
		}
	}
	return append(args, "--", r.URL)
}
