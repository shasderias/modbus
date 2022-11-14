package projectroot

import (
	"path"
	"runtime"
)

func Get() string {
	_, filename, _, _ := runtime.Caller(0)
	return path.Join(path.Dir(filename), "..", "..")
}
