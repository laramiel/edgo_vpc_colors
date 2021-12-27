// +build !windows

package watch

import (
	"os"
)

func MyOpenFile(name string) (file *os.File, err error) {
	return os.Open(name)
}
