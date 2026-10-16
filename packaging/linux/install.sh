#!/bin/sh
set -eu
app_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
data_dir="${XDG_DATA_HOME:-$HOME/.local/share}"
mkdir -p "$data_dir/applications" "$data_dir/icons/hicolor/scalable/apps"
cp "$app_dir/curb.svg" "$data_dir/icons/hicolor/scalable/apps/curb.svg"
{ printf '[Desktop Entry]\nType=Application\nName=Curb\n'; printf 'Exec=env GDK_BACKEND=wayland,x11 "%s/curb"\n' "$app_dir"; printf 'Icon=curb\nStartupWMClass=dev.curb.app\nTerminal=false\nCategories=AudioVideo;Utility;\n'; } > "$data_dir/applications/dev.curb.app.desktop"
printf 'Curb added to your application menu. Keep this folder in place.\n'
