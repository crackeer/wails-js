package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	rollRender "github.com/unrolled/render"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed frontend
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

func main() {
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:             "wails-js",
		Width:             1024,
		Height:            768,
		MinWidth:          1024,
		MinHeight:         768,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		//Assets:            assets,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: NewGinEngine(),
		},
		Menu:     nil,
		Logger:   nil,
		LogLevel: logger.DEBUG,

		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnBeforeClose:    app.beforeClose,
		OnShutdown:       app.shutdown,
		WindowStartState: options.Normal,
		Bind: []interface{}{
			app,
		},
		// Windows platform specific options
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			// DisableFramelessWindowDecorations: false,
			WebviewUserDataPath: "",
		},
		// Mac platform specific options
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: false,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 true,
				HideToolbarSeparator:       true,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			About: &mac.AboutInfo{
				Title:   "wails-js",
				Message: "",
				Icon:    icon,
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}

// NewGinEngine
//
//	@return *gin.Engine
func NewGinEngine() *gin.Engine {
	router := gin.Default()
	router.NoRoute(createStaticHandler("layout"))
	return router
}

func createStaticHandler(layout string) gin.HandlerFunc {
	object := rollRender.New(rollRender.Options{
		Directory:  "frontend",
		FileSystem: &rollRender.LocalFileSystem{},
		Layout:     layout,            // Specify a layout template. Layouts can call {{ yield }} to render the current template or {{ partial "css" }} to render a partial from the current template.
		Extensions: []string{".tmpl"}, // Specify extensions to load for templates.
		Delims: rollRender.Delims{
			Left:  "{[{",
			Right: "}]}",
		},
		IsDevelopment:               true,
		RenderPartialsWithoutPrefix: true,
	})

	return func(ctx *gin.Context) {
		file := strings.TrimLeft(ctx.Request.URL.Path, "/")
		fmt.Println(ctx.Request.URL.Path)
		file = strings.TrimRight(file, "/")
		if len(file) < 1 {
			file = "index"
		}

		var data interface{}
		jsonFile := file + ".json"
		if bytes, err := assets.ReadFile(jsonFile); err == nil {
			json.Unmarshal(bytes, &data)
		}

		object.HTML(ctx.Writer, http.StatusOK, file, data)
	}
}

func getPageNames(path string, ext string) []string {
	retData := []string{}
	list, err := assets.ReadDir(path)
	if err != nil {
		return retData
	}
	for _, item := range list {
		if item.IsDir() {
			tmpList := getPageNames(path+"/"+item.Name(), ext)
			retData = append(retData, tmpList...)
		} else if strings.HasSuffix(item.Name(), ext) {
			retData = append(retData, path+"/"+item.Name())
		}
	}
	return retData
}

type MockFileSystem struct {
	Asset embed.FS
}

func (fs *MockFileSystem) Open(name string) (fs.File, error) {
	return fs.Asset.Open(name)
}

func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Asset: assets,
	}
}
