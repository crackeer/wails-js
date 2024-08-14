package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

//go:embed frontend/public
var staticAssets embed.FS

//go:embed frontend/pages
var pageAssets embed.FS

//go:embed build/appicon.png
var icon []byte

func main() {
	app := NewApp()
	pageLocalPath := os.Getenv("PAGE_ASSETS_DIR")
	staticLocalPath := os.Getenv("STATIC_ASSETS_DIR")

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
			Assets:  NewMockFileSystem(staticAssets, staticLocalPath),
			Handler: NewGinEngine(pageAssets, pageLocalPath),
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
			WebviewUserDataPath:  "",
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
func NewGinEngine(embedFs embed.FS, localPath string) *gin.Engine {
	router := gin.Default()

	option := rollRender.Options{
		Directory:  localPath,
		FileSystem: rollRender.LocalFileSystem{},
		Layout:     "layout",
		Extensions: []string{".tmpl"}, // Specify extensions to load for templates.
		Delims: rollRender.Delims{
			Left:  "{[{",
			Right: "}]}",
		},
		IsDevelopment:               true,
		RenderPartialsWithoutPrefix: true,
	}
	if len(localPath) > 0 {
		option.Directory = localPath
		option.FileSystem = &rollRender.LocalFileSystem{}
	} else {
		option.Directory = "frontend/pages"
		option.FileSystem = &rollRender.EmbedFileSystem{embedFs}
	}
	rollRenderer := rollRender.New(option)

	router.NoRoute(func(ctx *gin.Context) {
		file := strings.TrimLeft(ctx.Request.URL.Path, "/")
		fmt.Println(ctx.Request.URL.Path)
		file = strings.TrimRight(file, "/")
		if len(file) < 1 {
			file = "index"
		}
		rollRenderer.HTML(ctx.Writer, http.StatusOK, file, nil)
	})
	return router
}

type MockFileSystem struct {
	Asset     embed.FS
	LocalPath string
}

func (fs *MockFileSystem) Open(name string) (fs.File, error) {
	if len(fs.LocalPath) > 0 {
		return os.Open(filepath.Join(fs.LocalPath, name))
	}
	return fs.Asset.Open(name)
}

// NewMockFileSystem
//
//	@param embedFs
//	@param localPath
//	@return *MockFileSystem
func NewMockFileSystem(embedFs embed.FS, localPath string) *MockFileSystem {
	return &MockFileSystem{
		Asset:     embedFs,
		LocalPath: localPath,
	}
}
