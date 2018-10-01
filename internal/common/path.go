package common

import (
	"os"
	"path"
)

// PathMkdirAll make a path by joining paths and create directory if not exists
func PathMkdirAll(paths ...string) string {
	target := path.Join(paths...)
	if _, err := os.Stat(target); os.IsNotExist(err) {
		_ = os.MkdirAll(target, 0711)
	}
	return target
}
