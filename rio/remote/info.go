package remote

import (
	"bytes"
	"fmt"
	"strings"

	"khan.rip/rio"
)

func (host *Host) Info() (*rio.Info, error) {
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

	info := &rio.Info{}
	info.Uname = o

	// try to make like GOOS
	info.OS = strings.ToLower(chunks[0])

	if info.OS == "openbsd" && len(chunks) > 4 {
		info.Hostname = chunks[1]
		info.Kernel = chunks[2]
		info.Arch = chunks[4]
	} else if info.OS == "linux" && len(chunks) > 4 {
		info.Hostname = chunks[1]
		info.Kernel = chunks[2]
		info.Arch = chunks[len(chunks)-2]
	} else {
		return nil, fmt.Errorf("Not sure how to parse uname -a: %#v", o)
	}

	// try to make like GOARCH
	if info.Arch == "x86_64" {
		info.Arch = "amd64"
	}

	return info, nil
}
