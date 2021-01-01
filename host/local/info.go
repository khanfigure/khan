package local

import (
	"os"
	"runtime"

	hhost "github.com/desops/khan/host"
)

func (host *Host) Info() (*hhost.Info, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &hhost.Info{
		Hostname: hostname,
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
	}, nil
}
