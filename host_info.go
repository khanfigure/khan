package khan

import (
	"bytes"
	"fmt"
	"strings"
)

func (host *Host) getInfo() error {
	host.VirtMu.Lock()
	defer host.VirtMu.Unlock()

	v := host.Virt

	if v.Hostname != "" {
		// already cached
		return nil
	}

	session, err := host.Run.Pool.Get(host.Host)
	if err != nil {
		return err
	}
	defer session.Put()

	outbuf := &bytes.Buffer{}

	session.Stdout = outbuf

	cmdline := "uname -a"

	if host.Run.Verbose {
		fmt.Println("ssh", host.Host, cmdline)
	}

	if err := session.Run(cmdline); err != nil {
		if host.Run.Verbose {
			fmt.Println("ssh", host.Host, cmdline, err)
		}

		return err
	}

	o := strings.TrimSpace(outbuf.String())
	chunks := strings.Split(o, " ")

	v.Uname = o
	v.OS = chunks[0]

	if v.OS == "OpenBSD" && len(chunks) > 4 {
		v.Hostname = chunks[1]
		v.Kernel = chunks[2]
		v.Arch = chunks[4]
		if v.Arch == "amd64" {
			// normalize just to make it easier
			v.Arch = "x86_64"
		}
	} else if v.OS == "Linux" && len(chunks) > 4 {
		v.Hostname = chunks[1]
		v.Kernel = chunks[2]
		v.Arch = chunks[len(chunks)-2]
	} else {
		return fmt.Errorf("Not sure how to parse uname -a: %#v", o)
	}

	return nil
}
