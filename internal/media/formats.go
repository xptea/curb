// Package media defines the output formats shared by settings, UI, and downloads.
package media

import "strings"

const (
	Original           = "Original"
	DefaultVideoOutput = "MP4"
	DefaultAudioOutput = "MP3"
)

var VideoOutputs = []string{Original, "MP4", "MKV"}
var AudioOutputs = []string{Original, "MP3", "M4A", "WAV", "FLAC"}

func VideoOutput(value string) string { return normalize(value, VideoOutputs, DefaultVideoOutput) }
func AudioOutput(value string) string { return normalize(value, AudioOutputs, DefaultAudioOutput) }

func normalize(value string, options []string, fallback string) string {
	for _, option := range options {
		if strings.EqualFold(strings.TrimSpace(value), option) {
			return option
		}
	}
	return fallback
}

func AudioHasBitrate(value string) bool {
	output := AudioOutput(value)
	return output == "MP3" || output == "M4A"
}
