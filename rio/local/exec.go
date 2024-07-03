package local

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"khan.rip/rio"
)

func (host *Host) Exec(cmd *rio.Cmd) error {
	if host.verbose {
		fmt.Println(host, cmd)
	}

	errbuf := &bytes.Buffer{}

	stderr := cmd.Stderr
	if stderr == nil {
		stderr = errbuf
	}

	c := exec.CommandContext(cmd.Context, cmd.Path, cmd.Args...)
	c.Stdin = cmd.Stdin
	c.Stdout = cmd.Stdout
	c.Stderr = stderr
	if len(cmd.Env) > 0 {
		// Maybe just change this to copy everything?
		c.Env = append(c.Env, "PATH="+os.Getenv("PATH"))
		for _, e := range cmd.Env {
			c.Env = append(c.Env, e[0]+"="+e[1])
		}
	}
	if err := c.Run(); err != nil {
		return &rio.CmdErr{Cmd: cmd, StdErr: strings.TrimSpace(errbuf.String()), ExecErr: err}
	}
	return nil
}
