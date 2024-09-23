package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"wails-js/bind"
	"wails-js/event"

	"github.com/gin-gonic/gin"
	rollRender "github.com/unrolled/render"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed frontend
var staticAssets embed.FS

var embedFsSubDir string = "frontend"

//go:embed build/T.png
var icon []byte

func main() {
	example := bind.NewExample()
	system := bind.NewSystem()
	assetsDir := os.Getenv("ASSETS_DIR")
	fmt.Println("assetsDir", assetsDir)

	// Create application with options
	err := wails.Run(&options.App{
		Title:             "开发者工具箱",
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
		AssetServer: &assetserver.Options{
			Handler: NewGinEngine(staticAssets, assetsDir),
		},
		Menu:     nil,
		Logger:   nil,
		LogLevel: logger.DEBUG,

		OnStartup: OnStartup,
		OnDomReady: func(ctx context.Context) {
			fmt.Println("Dom ready!")
		},
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			fmt.Println("OnBeforeClose")
			return false
		},
		OnShutdown: func(ctx context.Context) {
			fmt.Println("OnShutdown")
		},
		WindowStartState: options.Normal,
		Bind: []interface{}{
			example, system,
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
				Title:   "开发者工具箱",
				Message: "开发者工具箱",
				Icon:    icon,
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}

// OnStartup
//
//	@param ctx
func OnStartup(ctx context.Context) {
	runtime.EventsOn(ctx, "open-json-file", event.JSONFileSelect(ctx, "open-json-file-callback"))
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
	var fileServer http.Handler
	// embedFs去掉前缀
	subFs, _ := fs.Sub(embedFs, embedFsSubDir)
	myFileSystem := NewMyFileSystem(subFs, localPath)

	if len(localPath) > 0 {
		option.Directory = localPath
		option.FileSystem = &rollRender.LocalFileSystem{}
		fileServer = http.StripPrefix("", http.FileServer(http.Dir(localPath)))
	} else {
		option.Directory = embedFsSubDir
		option.FileSystem = &rollRender.EmbedFileSystem{FS: embedFs}
		fileServer = http.StripPrefix("", http.FileServer(http.FS(subFs)))
	}
	rollRenderer := rollRender.New(option)

	if len(localPath) > 0 {
		fileServer = http.StripPrefix("", http.FileServer(http.Dir(localPath)))
	}
	router.NoRoute(func(ctx *gin.Context) {
		file := strings.TrimLeft(ctx.Request.URL.Path, "/")
		if len(file) > 0 {
			if _, err := myFileSystem.Open(file); err == nil {
				ctx.Writer.Header().Set("Content-Type", myFileSystem.ContentType(ctx.Request.URL.Path))
				fileServer.ServeHTTP(ctx.Writer, ctx.Request)
				return
			}
		}

		file = strings.TrimRight(file, "/")
		if len(file) < 1 {
			file = "index"
		}
		var binding map[string]interface{} = map[string]interface{}{
			"file": file,
		}
		rollRenderer.HTML(ctx.Writer, http.StatusOK, file, binding)
	})
	return router
}

// MyFileSystem
type MyFileSystem struct {
	Asset     fs.FS
	LocalPath string
}

func (fs *MyFileSystem) Open(name string) (fs.File, error) {
	if len(fs.LocalPath) > 0 {
		return os.Open(filepath.Join(fs.LocalPath, name))
	}
	return fs.Asset.Open(name)
}

var mimeTypeByExtension = map[string]string{
	".css":  "text/css",
	".js":   "application/javascript",
	".html": "text/html",
	".svg":  "image/svg+xml",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".ico":  "image/x-icon",
	".txt":  "text/plain",
}

func (fs *MyFileSystem) ContentType(name string) string {
	extension := path.Ext(name)
	if mimeType, ok := mimeTypeByExtension[extension]; ok {
		return mimeType
	}
	return "application/octet-stream"
}

// NewMyFileSystem
//
//	@param embedFs
//	@param localPath
//	@return *MyFileSystem
func NewMyFileSystem(tmpFs fs.FS, localPath string) *MyFileSystem {
	return &MyFileSystem{
		Asset:     tmpFs,
		LocalPath: localPath,
	}
}
