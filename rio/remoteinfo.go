package rio

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
)

var (
	remoteinfocache   map[string]*remoteinfo
	remoteinfocachemu sync.Mutex
)

type remoteinfo struct {
	uname    string
	hostname string
	kernel   string
	os       string
	arch     string
}

func (config *Config) getremoteinfo(host string) (*remoteinfo, error) {
	remoteinfocachemu.Lock()
	defer remoteinfocachemu.Unlock()

	if remoteinfocache == nil {
		remoteinfocache = map[string]*remoteinfo{}
	}

	v, ok := remoteinfocache[host]
	if ok {
		return v, nil
	}

	session, err := config.Pool.Get(host)
	if err != nil {
		return nil, err
	}
	defer session.Put()

	outbuf := &bytes.Buffer{}

	session.Stdout = outbuf

	cmdline := "uname -a"

	if config.Verbose {
		fmt.Println("ssh", config.Host, cmdline)
	}

	if err := session.Run(cmdline); err != nil {
		if config.Verbose {
			fmt.Println("ssh", config.Host, cmdline, err)
		}

		return nil, err
	}

	o := strings.TrimSpace(outbuf.String())
	chunks := strings.Split(o, " ")

	ri := &remoteinfo{}
	ri.uname = o
	ri.os = chunks[0]
	if ri.os == "OpenBSD" && len(chunks) > 4 {
		ri.hostname = chunks[1]
		ri.kernel = chunks[2]
		ri.arch = chunks[4]
		if ri.arch == "amd64" {
			// normalize just to make it easier
			ri.arch = "x86_64"
		}
	} else if ri.os == "Linux" && len(chunks) > 4 {
		ri.hostname = chunks[1]
		ri.kernel = chunks[2]
		ri.arch = chunks[len(chunks)-2]
	} else {
		return nil, fmt.Errorf("Not sure how to parse uname -a: %#v", o)
	}

	remoteinfocache[host] = ri
	return ri, nil
}
