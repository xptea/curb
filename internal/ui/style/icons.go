package style

import (
	"github.com/egoist/mygo/ui"
)

var (
	ArrowIcon    = icon(`<path d="M5 12h14m-5-5 5 5-5 5"/>`)
	ChevronIcon  = icon(`<path d="m6 9 6 6 6-6"/>`)
	CheckIcon    = icon(`<path d="m5 12 4 4L19 6"/>`)
	SunIcon      = icon(`<circle cx="12" cy="12" r="4"/><path d="M12 2v2m0 16v2M2 12h2m16 0h2M4.93 4.93l1.42 1.42m11.3 11.3 1.42 1.42M4.93 19.07l1.42-1.42m11.3-11.3 1.42-1.42"/>`)
	MoonIcon     = icon(`<path d="M20.9 13A9 9 0 0 1 11 3.1 9 9 0 1 0 20.9 13Z"/>`)
	SettingsIcon = icon(`<path d="m9.5 3-.5 2-2 1-2-.5-2 3.5L4.5 11v2L3 15l2 3.5 2-.5 2 1 .5 2h5l.5-2 2-1 2 .5 2-3.5-1.5-2v-2L21 9l-2-3.5-2 .5-2-1-.5-2Z"/><circle cx="12" cy="12" r="3"/>`)
	MinimizeIcon = icon(`<path d="M5 12h14"/>`)
	MaximizeIcon = icon(`<rect x="5" y="5" width="14" height="14" rx="1"/>`)
	RestoreIcon  = icon(`<path d="M8 8V4h12v12h-4"/><rect x="4" y="8" width="12" height="12" rx="1"/>`)
	CloseIcon    = icon(`<path d="m6 6 12 12M6 18 18 6"/>`)
	BackIcon     = icon(`<path d="M19 12H5m5-5-5 5 5 5"/>`)
	VideoIcon    = icon(`<rect x="3" y="5" width="18" height="14" rx="2"/><path d="m10 9 5 3-5 3V9Z"/>`)
	AudioIcon    = icon(`<path d="M9 18V5l11-2v13M9 8l11-2"/><ellipse cx="6" cy="18" rx="3" ry="3"/><ellipse cx="17" cy="16" rx="3" ry="3"/>`)
)

func icon(paths string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round">` + paths + `</svg>`))
}
