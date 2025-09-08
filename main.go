package main

import (
	"embed"
	"os"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 禁用 GPU 加速
	if runtime.GOOS == "linux" {
		os.Setenv("WEBKIT_DISABLE_COMPOSITING_MODE", "1")
		os.Setenv("GDK_BACKEND", "x11")
		os.Setenv("WEBKIT_FORCE_SANDBOX", "0")
		os.Setenv("GDK_SCALE", "1")
		os.Setenv("GDK_DPI_SCALE", "1")
	}

	// Create an instance of the app structure
	app := CreateApp(assets)

	// Create application with options
	err := wails.Run(&options.App{
		Title:  APP_NAME,
		Width:  1180,
		Height: 750,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		// Frameless:        true,
		BackgroundColour:         &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		EnableDefaultContextMenu: false,
		Windows: &windows.Options{
			IsZoomControlEnabled: false, // 页面缩放比例
		},
		Linux: &linux.Options{
			WebviewGpuPolicy: linux.WebviewGpuPolicyNever, // or OnDemand/Always
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: true,
			CSSDropProperty:    "--wails-drop-target",
			CSSDropValue:       "drop",
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: func() string {
				return Env.ExePath
			}(),
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}

	// app.db.Close()
}
