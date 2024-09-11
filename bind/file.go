package bind

import (
	"os"
	"path/filepath"
)

// File
type File struct {
}

// NewFile ...
//
//	@return *Example
func NewFile() *File {
	return &File{}
}

// ListDir
//
//	@receiver a
//	@param dir
//	@return []map
func (a *File) ListDir(dir string) []map[string]interface{} {
	retData := []map[string]interface{}{}
	list, err := os.ReadDir(dir)
	if err != nil {
		return retData
	}

	for _, file := range list {
		info, err := file.Info()
		tmp := map[string]interface{}{
			"name":   file.Name(),
			"is_dir": file.IsDir(),
			"path":   filepath.Join(dir, file.Name()),
			"size":   0,
			"modify": info.ModTime().Format("2006-01-02 15:04:05"),
		}
		if err == nil {
			tmp["size"] = info.Size()
		}
		retData = append(retData, tmp)
	}
	return retData
}
