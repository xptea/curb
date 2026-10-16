package download

import (
	"strings"
	"testing"
)

func TestProgressParsing(t *testing.T) {
	u, ok := parseUpdate(`curb-progress:{"status":"downloading","downloaded_bytes":500,"total_bytes":1000,"speed":2048,"eta":5}`)
	if !ok || u.Progress != .5 || !strings.Contains(u.Detail, "2.0 KB/s") {
		t.Fatalf("wrong progress: %+v", u)
	}
	u, ok = parseUpdate(`curb-progress:{"status":"downloading","downloaded_bytes":20,"total_bytes_estimate":100,"speed":null,"eta":null}`)
	if !ok || u.Progress != .2 {
		t.Fatalf("estimate not used: %+v", u)
	}
	u, ok = parseUpdate(`curb-progress:{"status":"finished","downloaded_bytes":100,"total_bytes":100}`)
	if !ok || u.Progress != -1 {
		t.Fatal("should show processing until after_move")
	}
	u, ok = parseUpdate(`curb-file:"/tmp/a file.mp4"`)
	if !ok || u.Path != "/tmp/a file.mp4" {
		t.Fatal("saved path not parsed")
	}
	if _, ok := parseUpdate(`curb-progress:broken`); ok {
		t.Fatal("malformed progress accepted")
	}
}
