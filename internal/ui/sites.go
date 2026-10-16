package view

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"

	"curb/internal/ui/components"
	"curb/internal/ui/style"
)

const supportedSitesURL = "https://github.com/yt-dlp/yt-dlp/blob/master/supportedsites.md"

func (a *screen) supportedSitesView(c *ui.Context) {
	ui.Column(c).Grow(1).MinHeight(0).Padding(20, 32, 44, 32).Children(func() {
		ui.Column(c).Width(620).MaxWidthPercent(100).Margin(0, ui.Auto).Grow(1).MinHeight(0).Gap(12).Children(func() {
			components.PageHeader(c, "Supported sites", func() { a.SupportedSitesOpen = false }, func() {
				services := c.Services()
				ui.ButtonBase(c).Label("View yt-dlp’s public list").Shrink(0).OnClick(func() {
					services.OpenURL(supportedSitesURL)
				}).Children(func() { ui.Text(c, "Public list ↗").FontSize(12) })
			})
			var sites []string
			ui.Row(c).AlignItems(ui.Center).Gap(12).Shrink(0).Children(func() {
				ui.TextInput(c, &a.SitesQuery).Label("Search supported sites").Placeholder("Search sites…").Height(38).Grow(1).MinWidth(0).FontSize(13)
				// Read after the input: MyGo applies edits during the frame.
				sites = filterSites(a.SupportedSites, a.SitesQuery)
				if !a.SitesLoading && a.SitesError == "" {
					ui.Text(c, fmt.Sprintf("%d of %d handlers", len(sites), len(a.SupportedSites))).FontSize(11).TextColor(c.Theme().TextMuted).Shrink(0).SingleLine()
				}
			})
			switch {
			case a.SitesLoading:
				ui.Text(c, "Loading supported sites…").FontSize(12).TextColor(c.Theme().TextMuted)
			case a.SitesError != "":
				ui.Text(c, a.SitesError).FontSize(12).TextColor(c.Theme().Danger)
				ui.Button(c, "Retry").AlignSelf(ui.Start).OnClick(a.actions.LoadSupportedSites)
			default:
				if len(sites) == 0 {
					ui.Text(c, "No matching sites. Unlisted links may still work.").FontSize(12).TextColor(c.Theme().TextMuted)
				} else {
					// A virtual list builds only visible rows, keeping resizing and
					// scrolling responsive even with thousands of handlers.
					ui.List(c.Key("supported-sites-list:"+a.SitesQuery), nil, len(sites), func(i int) {
						ui.Text(c, sites[i]).FontSize(12).Padding(7, 10)
					}).Label("Supported site handlers").Grow(1).MinHeight(0).Border(1, c.Theme().Border).Radius(style.Radius)
				}
			}
		})
	})
}

func filterSites(sites []string, query string) []string {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return sites
	}
	var matches []string
	for _, site := range sites {
		if strings.Contains(strings.ToLower(site), query) {
			matches = append(matches, site)
		}
	}
	return matches
}
