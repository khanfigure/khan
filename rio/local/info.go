package local

import (
	"os"
	"runtime"

	"github.com/desops/khan/rio"
)

func (host *Host) Info() (*rio.Info, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &rio.Info{
		Hostname: hostname,
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
	}, nil
}
