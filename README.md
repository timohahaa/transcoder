# Split and stitch video transcoder
This is my bachelor's thesis at ETU :)

### Notes on encoder
I use `systemd-run` to pin ffmpeg processes to specific CPU-cores. It is done this way to better use system resources for an encoder. Because of this, you can't really run encoder service inside Docker. There is no solution that I know of, and even if there is, I'm not really bothered to look for one as I hava a bare-metal Linux systemd I can test this service on :)

### Potential optimizations
- use less `ffprobe` calls :_)
