package remote

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/desops/khan/rio"

	"github.com/keegancsmith/shell"
)

func (host *Host) Exec(cmd *rio.Cmd) error {
	fmt.Println(host, cmd)

	errbuf := &bytes.Buffer{}

	stderr := cmd.Stderr
	if stderr == nil {
		stderr = errbuf
	}

	session, err := host.pool.Get(host.connect)
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdin = cmd.Stdin
	session.Stdout = cmd.Stdout
	session.Stderr = stderr

	if !cmd.Shell {
		// This doesn't often work-- you would need to set AcceptENV on /etc/ssh/sshd_config
		// on the server with each var you allow to set. I'm not sure why things are that way.
		// In shell mode we can work around it with some export statements.
		for _, e := range cmd.Env {
			if err := session.Setenv(e[0], e[1]); err != nil {
				return err
			}
		}
	}

	cmdline := cmd.Path
	for _, a := range cmd.Args {
		cmdline += " " + shell.ReadableEscapeArg(a)
	}

	if cmd.Shell {
		exports := "source /etc/profile; "
		for _, e := range cmd.Env {
			exports += "export " + shell.ReadableEscapeArg(e[0]) + "=" + shell.ReadableEscapeArg(e[1]) + "; "
		}
		cmdline = "bash -c " + shell.ReadableEscapeArg(exports+cmdline)
	}

	err = session.Run(cmdline)

	if err != nil {
		return &rio.CmdErr{Cmd: cmd, StdErr: strings.TrimSpace(errbuf.String()), ExecErr: err}
	}
	return nil
}
