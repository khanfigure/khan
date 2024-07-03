package khan

import (
	"fmt"
	"io"
	"runtime"

	"khan.rip/rio"
)

// Host is the context for an execution run on a specific server. It contains a virtual model
// of all changes we are capable of making on a server, including service status and a filesystem.
// This way a dry run can be pixel-perfect.
type Host struct {
	Run *Run

	Verbose bool

	Name string // Friendly name for host
	SSH  bool
	Host string // Host for SSH

	rh rio.Host
}

func (host *Host) Key() string {
	if host.SSH {
		return host.Host
	}
	return "local"
}
func (host *Host) String() string {
	return host.rh.String()
}

func (host *Host) Add(add ...Item) error {
	_, fn, line, _ := runtime.Caller(1)
	source := fmt.Sprintf("%s:%d", fn, line)
	return host.AddFromSource(source, add...)
}

func (host *Host) AddFromSource(source string, add ...Item) error {
	host.Run.itemsmu.Lock()
	defer host.Run.itemsmu.Unlock()

	for _, item := range add {
		if err := host.Run.addHostItem(host, source, item); err != nil {
			return err
		}
	}
	return nil
}

func (host *Host) OS() (string, error) {
	info, err := host.rh.Info()
	if err != nil {
		return "", err
	}
	return info.OS, nil
}

// Passthrough functions to rio.host
func (host *Host) Open(path string) (io.ReadCloser, error) {
	return host.rh.Open(path)
}
