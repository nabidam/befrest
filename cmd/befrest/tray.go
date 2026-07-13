package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"log/slog"
	"os"
	"runtime"

	"fyne.io/systray"
)

func trayAvailable() bool {
	return runtime.GOOS != "linux" || os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
}

func startTray(openBefrest, openLog, quit func()) {
	go systray.Run(func() {
		if icon, err := trayIcon(); err == nil {
			systray.SetIcon(icon)
		}
		systray.SetTitle("Befrest")
		systray.SetTooltip("Befrest local sharing hub")

		openItem := systray.AddMenuItem("Open befrest", "Open Befrest in your browser")
		logItem := systray.AddMenuItem("Open log", "Open the Befrest log")
		quitItem := systray.AddMenuItem("Quit", "Stop Befrest")
		go func() {
			for {
				select {
				case <-openItem.ClickedCh:
					openBefrest()
				case <-logItem.ClickedCh:
					openLog()
				case <-quitItem.ClickedCh:
					quit()
					systray.Quit()
					return
				}
			}
		}()
	}, func() { slog.Info("tray stopped") })
}

func trayIcon() ([]byte, error) {
	const size = 16
	image := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if x == 0 || y == 0 || x == size-1 || y == size-1 {
				image.SetNRGBA(x, y, color.NRGBA{R: 14, G: 165, B: 233, A: 255})
				continue
			}
			image.SetNRGBA(x, y, color.NRGBA{R: 8, G: 47, B: 73, A: 255})
		}
	}
	var encoded bytes.Buffer
	err := png.Encode(&encoded, image)
	return encoded.Bytes(), err
}
