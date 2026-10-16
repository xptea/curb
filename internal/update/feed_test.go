package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func fixtureClient(t *testing.T, platform string, manifest Manifest, payload []byte) *Client {
	t.Helper()
	data, _ := json.Marshal(manifest)
	client := NewClient("0.1.0")
	client.Platform = platform
	client.HTTP.Transport = transportFunc(func(request *http.Request) (*http.Response, error) {
		body := payload
		if request.URL.String() == FeedURL {
			body = data
		} else if !strings.HasPrefix(request.URL.String(), ReleaseURL) {
			t.Errorf("unexpected URL: %s", request.URL)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(body))), ContentLength: int64(len(body)), Header: make(http.Header)}, nil
	})
	return client
}

func testAsset(payload []byte, name, format string) Asset {
	sum := sha256.Sum256(payload)
	return Asset{Name: name, SHA256: hex.EncodeToString(sum[:]), Size: int64(len(payload)), Format: format}
}

func TestCheckSelectsNewerPlatformRelease(t *testing.T) {
	asset := testAsset([]byte("release"), "Curb-0.2.0-darwin-universal.tar.gz", "tar.gz")
	manifest := Manifest{Schema: 1, Version: "0.2.0", Platforms: map[string]Asset{"darwin-universal": asset}}
	for _, platform := range []string{"darwin-amd64", "darwin-arm64", "linux-amd64"} {
		client := fixtureClient(t, platform, manifest, nil)
		release, err := client.Check(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if platform == "linux-amd64" {
			if release != nil {
				t.Fatal("offered macOS update to Linux")
			}
			continue
		}
		if release == nil || release.Platform != "darwin-universal" || release.Version != "0.2.0" {
			t.Fatalf("wrong platform release: %+v", release)
		}
		if release.URL() != ReleaseURL+"v0.2.0/"+asset.Name {
			t.Fatal("release URL is not pinned to the expected repository/tag")
		}
		client.Current = "0.2.0"
		if same, err := client.Check(context.Background()); err != nil || same != nil {
			t.Fatal("offered same version", err)
		}
		client.Current = "1.0.0"
		if older, err := client.Check(context.Background()); err != nil || older != nil {
			t.Fatal("offered downgrade", err)
		}
	}
}

func TestVersionComparison(t *testing.T) {
	for _, test := range []struct {
		next, current string
		newer         bool
	}{
		{"0.10.0", "0.9.9", true}, {"1.0.0", "0.99.99", true}, {"1.0.0", "1.0.0", false}, {"0.1.9", "0.2.0", false},
	} {
		got, err := newerVersion(test.next, test.current)
		if err != nil || got != test.newer {
			t.Fatalf("%s > %s: %v, %v", test.next, test.current, got, err)
		}
	}
	for _, invalid := range []string{"", "v1.0.0", "1.2", "01.0.0", "1.0.0-beta", "1.0.0/../../bad"} {
		if _, err := parseVersion(invalid); err == nil {
			t.Fatalf("accepted %q", invalid)
		}
	}
}

func TestReleaseValidation(t *testing.T) {
	asset := testAsset([]byte("release"), "Curb.AppImage", "appimage")
	for _, mutate := range []func(*Release){
		func(r *Release) { r.Asset.Name = "../other.AppImage" },
		func(r *Release) { r.Asset.Name = "bad\\other.AppImage" },
		func(r *Release) { r.Asset.SHA256 = "wrong" },
		func(r *Release) { r.Asset.Size = 0 },
		func(r *Release) { r.Asset.Format = "zip" },
	} {
		release := Release{Version: "0.2.0", Platform: "linux-amd64", Asset: asset}
		mutate(&release)
		if release.validate() == nil {
			t.Fatalf("accepted bad release: %+v", release)
		}
	}
}

func TestDownloadRejectsCorruptAndIncompletePayloads(t *testing.T) {
	expected := []byte("verified release")
	release := Release{Version: "0.2.0", Platform: "linux-amd64", Asset: testAsset(expected, "Curb.AppImage", "appimage")}
	for _, body := range [][]byte{[]byte("verified releasX"), []byte("truncated"), expected} {
		client := fixtureClient(t, release.Platform, Manifest{}, body)
		destination := filepath.Join(t.TempDir(), "asset")
		err := client.download(context.Background(), release, destination, nil)
		if string(body) == string(expected) {
			if err != nil {
				t.Fatal(err)
			}
			data, _ := os.ReadFile(destination)
			if string(data) != string(expected) {
				t.Fatal("download changed")
			}
		} else if err == nil {
			t.Fatal("accepted corrupt payload")
		}
	}
}

func TestFailedPreparationKeepsInstalledApp(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "Curb.AppImage")
	if err := os.WriteFile(target, []byte("current app"), 0o755); err != nil {
		t.Fatal(err)
	}
	payload := []byte("bad payload")
	release := Release{Version: "0.2.0", Platform: "linux-amd64", Asset: testAsset(payload, "Curb.AppImage", "appimage")}
	client := fixtureClient(t, release.Platform, Manifest{}, payload)
	if _, err := client.Prepare(context.Background(), release, Installation{Kind: "appimage", Target: target, Platform: release.Platform}, nil); err == nil {
		t.Fatal("accepted a non-AppImage")
	}
	data, _ := os.ReadFile(target)
	if string(data) != "current app" {
		t.Fatal("failed update changed installed app")
	}
	entries, _ := os.ReadDir(root)
	if len(entries) != 1 {
		t.Fatal("failed update left staging files")
	}
}

func TestManagedFolderRefusesUnrelatedFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "Curb.exe"), []byte("app"), 0o755); err != nil {
		t.Fatal(err)
	}
	marker := installMarker{Product: product, Version: "0.1.0", Platform: "windows-amd64", Files: []string{markerName, "Curb.exe"}}
	data, _ := json.Marshal(marker)
	if err := os.WriteFile(filepath.Join(root, markerName), data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readMarker(root, marker.Platform, marker.Version); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "personal-video.mp4"), []byte("keep me"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readMarker(root, marker.Platform, marker.Version); err == nil {
		t.Fatal("accepted a folder containing unrelated user data")
	}
	if _, err := os.Stat(filepath.Join(root, "personal-video.mp4")); err != nil {
		t.Fatal("modified unrelated data")
	}
}
