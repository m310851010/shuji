package main

import (
    "context"
	"embed"
 	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
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
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		EnableDefaultContextMenu: false,
		Windows: &windows.Options{
        			IsZoomControlEnabled: false, // 页面缩放比例
        		},
        DragAndDrop: &options.DragAndDrop{
          EnableFileDrop:       true,
          DisableWebViewDrop:   true,
          CSSDropProperty:      "--wails-drop-target",
          CSSDropValue:         "drop",
        },
        SingleInstanceLock: &options.SingleInstanceLock{
                    UniqueId: func() string {
                        return Env.ExePath
                    }(),
        },
		OnStartup:        app.startup,
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
        			runtime.EventsEmit(ctx, "onBeforeClose")
//         			 dialog, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
//                             Type:          runtime.QuestionDialog,
//                             Title:         "Quit?",
//                             Message:       "Are you sure you want to quit?",
//                         })
//
//                         if err != nil {
//                             return false
//                         }
                        return true
        		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}

	// app.db.Close()
}
