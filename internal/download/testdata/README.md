# Media fixture

`sample.mp4` is a one-second green frame with a generated sine tone. It contains no downloaded or copyrighted source media. It was generated with:

```sh
ffmpeg -v error -f lavfi -i color=c=green:s=160x90:r=15 \
  -f lavfi -i sine=frequency=440:sample_rate=44100 -t 1 \
  -c:v libx264 -c:a aac -shortest sample.mp4
```

`sample.webm` contains the same generated scene encoded as VP9/Opus, to test conversion into a compatible MP4 and preservation of original WebM/Opus media:

```sh
ffmpeg -v error -i sample.mp4 -c:v libvpx-vp9 -b:v 80k -c:a libopus -b:a 48k sample.webm
```

Generating these fixtures requires a full development FFmpeg. Running the tests uses the checked-in fixtures and Curb's bundled conversion tools. Production includes the selected output encoders, but does not need synthetic media generators or VP9/Opus encoders.
