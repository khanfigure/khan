package rio

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/keegancsmith/shell"
)

func (config *Config) Stat(host string, sudo bool, path string) (os.FileInfo, error) {
	if config.Pool == nil {
		return os.Stat(path)
	}

	session, err := config.Pool.Get(host)
	if err != nil {
		return nil, err
	}

	outbuf := &bytes.Buffer{}
	errbuf := &bytes.Buffer{}

	session.Stdout = outbuf
	session.Stderr = errbuf

	cmd := "stat " + shell.ReadableEscapeArg(path)
	if sudo {
		cmd = "sudo " + cmd
	}
	//fmt.Println(cmd)
	if err := session.Run(cmd); err != nil {
		//fmt.Println(cmd, err)

		e := strings.TrimSpace(errbuf.String())

		if strings.HasPrefix(e, "stat: ") && strings.HasSuffix(e, "No such file or directory") {
			// emulate os.Stat
			return nil, &os.PathError{
				Op:   "stat",
				Path: path,
				Err:  syscall.ENOENT,
			}
		}

		return nil, err
	}

	fmt.Println("STAT output:", outbuf.String())
	return nil, fmt.Errorf("fixme")
}
