package remote

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/desops/khan/rio"

	"github.com/keegancsmith/shell"
)

func (host *Host) Exec(cmd *rio.Cmd) error {
	if !cmd.ReadOnly {
		fmt.Println(host, cmd)
	}

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
		// Capture certain stderr responses for programs like rm, stat, chmod, chown, etc
		// and emulate the kind of error you would get from the "os" package if the file
		// did not exist. For this to work properly, you need to always specficy the file
		// as the last argument. (i.e. Don't go "rm (file) -f" or the PathError will have "-f"
		// as the Path element.)
		// TODO put hack here for "cp (src) (dest)" so the PathError would correctly use (src).
		e := strings.TrimSpace(errbuf.String())

		if strings.HasPrefix(e, cmd.Path+": ") && len(cmd.Args) > 0 && strings.HasSuffix(e, "No such file or directory") {
			return &os.PathError{
				Op:   cmd.Path,
				Path: cmd.Args[len(cmd.Args)-1],
				Err:  syscall.ENOENT,
			}
		}

		// Otherwise bundle stderr into the error for display
		return &rio.CmdErr{Cmd: cmd, StdErr: e, ExecErr: err}
	}
	return nil
}
