# markit

markit is a simple tool to watermark images, on desktop and android devices.

Currently, watermarks must be in png format, and images in jpg.

# Building

markit is written in Go, using the Fyne toolkit. To build, you will need a working Go compiler with CGO, and the dependencies for Fyne.
Please see https://developer.fyne.io/started/#prerequisites for those.

Once you have them, installing is as simple as:

```
go install github.com/wltechblog/markit@latest
```

Android builds and other cross compiling are done with fyne-cross, see https://github.com/fyne-io/fyne-cross
