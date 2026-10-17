package media

import "testing"

func TestOutputDefaultsAndNormalization(t *testing.T) {
	for _, value := range []string{"", "unsupported", "--exec bad"} {
		if VideoOutput(value) != "MP4" || AudioOutput(value) != "MP3" {
			t.Fatalf("unsafe or missing defaults for %q", value)
		}
	}
	if VideoOutput(" original ") != Original || AudioOutput("flac") != "FLAC" {
		t.Fatal("format normalization failed")
	}
	for _, format := range AudioOutputs {
		want := format == "MP3" || format == "M4A"
		if AudioHasBitrate(format) != want {
			t.Fatalf("incorrect bitrate control for %s", format)
		}
	}
}
