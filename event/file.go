package event

import (
	"context"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// JSONFileSelect
//
//	@param ctx
//	@return ...interface{}
//	@return func( ...interface{})
func JSONFileSelect(ctx context.Context, callbackEvent string) func(...interface{}) {
	return func(optionalData ...interface{}) {
		fileName, err := runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
			Filters: []runtime.FileFilter{
				{DisplayName: "*.json", Pattern: "*.json"},
			},
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		bytes, err := os.ReadFile(fileName)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(bytes), callbackEvent)
		runtime.EventsEmit(ctx, callbackEvent, map[string]interface{}{
			"code": 0,
			"data": string(bytes),
		})
	}
}
