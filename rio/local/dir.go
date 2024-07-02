package local

import (
	"os"
)

func (host *Host) MkdirAll(fpath string) error {
	return os.MkdirAll(fpath, 0700)
}
