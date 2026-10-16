package download

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Update struct {
	Status, Detail, Path string
	Progress             float64
}

func parseUpdate(line string) (Update, bool) {
	u := Update{Progress: -1}
	if raw, ok := strings.CutPrefix(line, "curb-title:"); ok {
		if json.Unmarshal([]byte(raw), &u.Status) == nil {
			u.Status = "Downloading " + u.Status
			return u, true
		}
	}
	if raw, ok := strings.CutPrefix(line, "curb-file:"); ok {
		if json.Unmarshal([]byte(raw), &u.Path) == nil {
			return u, true
		}
	}
	if raw, ok := strings.CutPrefix(line, "curb-progress:"); ok {
		var p struct {
			Status     string  `json:"status"`
			Downloaded float64 `json:"downloaded_bytes"`
			Total      float64 `json:"total_bytes"`
			Estimate   float64 `json:"total_bytes_estimate"`
			Speed      float64 `json:"speed"`
			ETA        float64 `json:"eta"`
		}
		if json.Unmarshal([]byte(raw), &p) != nil {
			return u, false
		}
		if p.Total == 0 {
			p.Total = p.Estimate
		}
		u.Status = "Downloading"
		if p.Total > 0 {
			u.Progress = min(1, max(0, p.Downloaded/p.Total))
			u.Status = fmt.Sprintf("Downloading · %.0f%%", u.Progress*100)
		}
		if p.Speed > 0 {
			u.Detail = fmt.Sprintf("%s/s", humanBytes(p.Speed))
		}
		if p.ETA > 0 {
			u.Detail += fmt.Sprintf(" · %s left", (time.Duration(p.ETA) * time.Second).String())
		}
		if p.Status == "finished" {
			u.Status, u.Detail, u.Progress = "Finishing your file…", "", -1
		}
		return u, true
	}
	return u, false
}

func humanBytes(b float64) string {
	for _, unit := range []string{"B", "KB", "MB", "GB"} {
		if b < 1024 || unit == "GB" {
			return fmt.Sprintf("%.1f %s", b, unit)
		}
		b /= 1024
	}
	return ""
}
