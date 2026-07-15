package main

import (
	_ "embed"
	"log/slog"
	"os"
	"runtime"

	"fyne.io/systray"
)

//go:embed assets/tray-icon.png
var trayIconPNG []byte

func trayAvailable() bool {
	return runtime.GOOS != "linux" || os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
}

func startTray(openBefrest, openLog, quit func()) {
	go systray.Run(func() {
		systray.SetIcon(trayIconPNG)
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
