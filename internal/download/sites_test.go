package download

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestSupportedSitesWithoutEngine(t *testing.T) {
	if _, err := SupportedSites(context.Background(), Tools{}); err == nil {
		t.Fatal("missing engine should return a recoverable error")
	}
}

func TestRealSupportedSites(t *testing.T) {
	tools := integrationTools(t)
	sites, err := SupportedSites(context.Background(), tools)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(sites, "youtube") || !slices.Contains(sites, "vimeo") || !slices.Contains(sites, "generic") {
		t.Fatalf("incomplete extractor list (%d entries)", len(sites))
	}
	for _, site := range sites {
		if strings.Contains(strings.ToLower(site), "(currently broken)") {
			t.Fatalf("broken handler is visible: %q", site)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := SupportedSites(ctx, tools); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled lookup: %v", err)
	}
}
