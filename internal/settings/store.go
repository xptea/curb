package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Preferences contains the persistent defaults, separate from transient UI state.
type Preferences struct {
	Destination       string `json:"destination"`
	AutoCopyDownloads bool   `json:"auto_copy_downloads"`
	VideoOutputFormat string `json:"video_output_format"`
	AudioOutputFormat string `json:"audio_output_format"`
}

// Path returns the per-user settings location without creating it.
func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "curb", "settings.json"), nil
}

func Load(path string) (Preferences, error) {
	var preferences Preferences
	data, err := os.ReadFile(path)
	if err != nil {
		return preferences, err
	}
	err = json.Unmarshal(data, &preferences)
	return preferences, err
}

// Save replaces settings atomically with a private temporary file.
func Save(path string, p Preferences) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".settings-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(append(data, '\n')); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
