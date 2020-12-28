package khan

import (
	"fmt"
	"runtime"
	"sync"
)

// Host is the context for an execution run on a specific server. It contains a virtual model
// of all changes we are capable of making on a server, including service status and a filesystem.
// This way a dry run can be pixel-perfect.
type Host struct {
	Run *Run

	Name string // Friendly name for host

	SSH bool

	Host string // Host for SSH

	VirtMu sync.RWMutex
	Virt   *Virtual // Virtual model of the host
}

func (host *Host) Key() string {
	if host.SSH {
		return host.Host
	}
	return "local"
}
func (host *Host) String() string {
	title := host.Name + " "
	if host.Host == "" {
		title += "(local mode)"
	} else {
		title += "(ssh"
		if host.Host != host.Name {
			title += " " + host.Host
		}
		title += ")"
	}
	return title
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
