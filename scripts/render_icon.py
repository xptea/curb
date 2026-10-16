"""Render resources/curb.svg to MyGo's master PNG (developer tool only).

Requires librsvg and Cairo, available as librsvg2-2 and libcairo2 on Linux.
The generated PNG is committed; normal builds need no icon-rendering tools.
"""
import ctypes as c
from ctypes.util import find_library
from pathlib import Path


def main():
    rsvg = c.CDLL(find_library("rsvg-2"))
    cairo = c.CDLL(find_library("cairo"))
    gobject = c.CDLL(find_library("gobject-2.0"))

    class Rectangle(c.Structure):
        _fields_ = [(name, c.c_double) for name in ("x", "y", "width", "height")]

    def bind(library, name, result, *arguments):
        function = getattr(library, name)
        function.restype, function.argtypes = result, arguments
        return function

    load = bind(rsvg, "rsvg_handle_new_from_file", c.c_void_p, c.c_char_p, c.c_void_p)
    render = bind(rsvg, "rsvg_handle_render_document", c.c_int,
                  c.c_void_p, c.c_void_p, c.POINTER(Rectangle), c.c_void_p)
    surface_new = bind(cairo, "cairo_image_surface_create", c.c_void_p, c.c_int, c.c_int, c.c_int)
    context_new = bind(cairo, "cairo_create", c.c_void_p, c.c_void_p)
    write_png = bind(cairo, "cairo_surface_write_to_png", c.c_int, c.c_void_p, c.c_char_p)
    context_free = bind(cairo, "cairo_destroy", None, c.c_void_p)
    surface_free = bind(cairo, "cairo_surface_destroy", None, c.c_void_p)
    unref = bind(gobject, "g_object_unref", None, c.c_void_p)
    resources = Path(__file__).resolve().parents[1] / "resources"
    handle = load(str(resources / "curb.svg").encode(), None)
    if not handle:
        raise RuntimeError("Could not load resources/curb.svg")
    surface = surface_new(0, 1024, 1024)  # CAIRO_FORMAT_ARGB32
    context = context_new(surface)
    try:
        if not render(handle, context, c.byref(Rectangle(0, 0, 1024, 1024)), None):
            raise RuntimeError("Could not render app icon")
        if write_png(surface, str(resources / "icon.png").encode()):
            raise RuntimeError("Could not write resources/icon.png")
    finally:
        context_free(context)
        surface_free(surface)
        unref(handle)
    print("Rendered resources/icon.png")


if __name__ == "__main__":
    main()
