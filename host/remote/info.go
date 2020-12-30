package remote

import (
	"bytes"
	"fmt"
	"strings"

	hhost "github.com/desops/khan/host"
)

func (host *Host) Info() (*hhost.Info, error) {
	host.infomu.Lock()
	defer host.infomu.Unlock()

	if host.info != nil {
		return host.info, nil
	}

	session, err := host.pool.Get(host.connect)
	if err != nil {
		return nil, err
	}
	defer session.Put()

	outbuf := &bytes.Buffer{}

	session.Stdout = outbuf

	cmdline := "uname -a"

	if err := session.Run(cmdline); err != nil {
		return nil, err
	}

	o := strings.TrimSpace(outbuf.String())
	chunks := strings.Split(o, " ")

	info := &hhost.Info{}

	info.Uname = o
	info.OS = chunks[0]

	if info.OS == "OpenBSD" && len(chunks) > 4 {
		info.Hostname = chunks[1]
		info.Kernel = chunks[2]
		info.Arch = chunks[4]
		if info.Arch == "amd64" {
			// normalize just to make it easier
			info.Arch = "x86_64"
		}
	} else if info.OS == "Linux" && len(chunks) > 4 {
		info.Hostname = chunks[1]
		info.Kernel = chunks[2]
		info.Arch = chunks[len(chunks)-2]
	} else {
		return nil, fmt.Errorf("Not sure how to parse uname -a: %#v", o)
	}

	return info, nil
}
