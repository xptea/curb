package platform

import (
	"log"
	"os"

	"github.com/ebitengine/purego"
	"github.com/egoist/mygo"
)

// Older Xlib releases require explicit thread initialization before GTK starts.
func PrepareApplication() {
	if os.Getenv("CURB_BUNDLED_RUNTIME") == "1" {
		// The ELF loader has cached AppRun's library path for this process.
		// Restore the environment inherited by external desktop applications.
		if original := os.Getenv("CURB_ORIGINAL_LD_LIBRARY_PATH"); original != "" {
			os.Setenv("LD_LIBRARY_PATH", original)
		} else {
			os.Unsetenv("LD_LIBRARY_PATH")
		}
	}
	lib, err := purego.Dlopen("libX11.so.6", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err == nil {
		var initThreads func() int32
		purego.RegisterLibFunc(&initThreads, lib, "XInitThreads")
		initThreads()
	}
	// GTK uses GLib's program name for the Wayland app_id and X11's
	// WM_CLASS instance. Match our desktop entry so launchers group the icon.
	if library, err := purego.Dlopen("libglib-2.0.so.0", purego.RTLD_NOW|purego.RTLD_GLOBAL); err == nil {
		var setName func(string)
		purego.RegisterLibFunc(&setName, library, "g_set_prgname")
		setName("dev.curb.app")
	}
}

// MyGo paints the rounded transparent window. Leave GTK's visual and input
// shapes unrestricted: resizing those regions can invalidate the native
// surface's clip and blank its content after maximize, and adds X11 work to
// every resize. The compositor handles the alpha channel of MyGo's surface.
func ConfigureWindow(win *mygo.Window) {
	gtk, err := purego.Dlopen("libgtk-3.so.0", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		log.Printf("Window corners: %v", err)
		return
	}
	var getStyle func(uintptr) uintptr
	var newProvider func() uintptr
	var loadCSS func(uintptr, string, int64, uintptr) bool
	var addProvider func(uintptr, uintptr, uint32)
	var unref func(uintptr)
	purego.RegisterLibFunc(&getStyle, gtk, "gtk_widget_get_style_context")
	purego.RegisterLibFunc(&newProvider, gtk, "gtk_css_provider_new")
	purego.RegisterLibFunc(&loadCSS, gtk, "gtk_css_provider_load_from_data")
	purego.RegisterLibFunc(&addProvider, gtk, "gtk_style_context_add_provider")
	purego.RegisterLibFunc(&unref, gtk, "g_object_unref")
	provider := newProvider()
	// Only the GTK window backing is transparent; all visible chrome is
	// drawn by the app.
	css := "window { background-color: transparent; background-image: none; }"
	loadCSS(provider, css, int64(len(css)), 0)
	addProvider(getStyle(win.NativeHandle()), provider, 600)
	unref(provider)
}
