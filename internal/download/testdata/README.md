# Media fixture

`sample.mp4` is a one-second green frame with a generated sine tone. It contains no downloaded or copyrighted source media. It was generated with:

```sh
ffmpeg -v error -f lavfi -i color=c=green:s=160x90:r=15 \
  -f lavfi -i sine=frequency=440:sample_rate=44100 -t 1 \
  -c:v libx264 -c:a aac -shortest sample.mp4
```

Generating the fixture requires a full development FFmpeg. Running the tests uses this checked-in fixture and Curb's small bundled tools; the production app does not need video encoders or synthetic media generators.
