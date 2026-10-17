#!/bin/sh
set -eu
# The source archive is verified against the SHA-256 pin in packaging/tools.json.
source_dir=$1
output_dir=$2
cd "$source_dir"
./configure \
    --prefix="$output_dir" --disable-autodetect --disable-doc --disable-debug \
    --disable-static --enable-shared --disable-programs --enable-ffmpeg --enable-ffprobe \
    --disable-avdevice --disable-hwaccels \
    --disable-encoders --enable-encoder=libmp3lame,libx264,aac,flac,pcm_s16le \
    --disable-decoders \
    --enable-decoder='aac,aac_fixed,ac3,eac3,alac,flac,mp3,mp3float,mp2,mp2float,opus,vorbis,wavpack,pcm*,adpcm*,wmav1,wmav2,wmapro,wmalossless,amrnb,amrwb,h264,hevc,vp8,vp9,av1' \
    --disable-filters --enable-filter=aresample,anull,atrim,asetpts,scale,format,null \
    --enable-gpl --enable-libx264 --enable-libmp3lame --enable-gnutls --enable-zlib \
    --extra-cflags=-Os
make -j8
make install
cp COPYING.GPLv2 "$output_dir/FFmpeg-LICENSE"
cp ffbuild/config.log "$output_dir/config.log"
