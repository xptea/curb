package download

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// SupportedSites asks the installed engine for its available extractor list.
// This works offline and stays in sync when the bundled engine is updated.
// Keep site-specific variants, but exclude handlers the engine marks as broken.
func SupportedSites(ctx context.Context, tools Tools) ([]string, error) {
	if tools.Downloader == "" {
		return nil, errors.New("The download engine is missing. You can still view yt-dlp’s public list.")
	}
	cmd := exec.CommandContext(ctx, tools.Downloader, "--ignore-config", "--no-plugin-dirs", "--list-extractors")
	configureProcess(cmd)
	output, err := cmd.Output()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil {
		return nil, fmt.Errorf("Could not load supported sites: %w. You can still view yt-dlp’s public list.", err)
	}
	var sites []string
	for _, line := range strings.Split(string(output), "\n") {
		site := strings.TrimSpace(line)
		if site == "" || strings.Contains(strings.ToLower(site), "(currently broken)") {
			continue
		}
		sites = append(sites, site)
	}
	if len(sites) == 0 {
		return nil, errors.New("The download engine returned an empty list. You can still view yt-dlp’s public list.")
	}
	sort.SliceStable(sites, func(i, j int) bool { return strings.ToLower(sites[i]) < strings.ToLower(sites[j]) })
	return sites, nil
}
