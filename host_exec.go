package khan

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/keegancsmith/shell"
)

type Cmd struct {
	Path string
	Args []string
	Env  [][2]string
	Dir  string

	Shell bool // when ssh-ing, try to start a shell instead of just executing a command (gives you working environment vars, etc)

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	Context context.Context

	host *Host
}

type CmdErr struct {
	Cmd     *Cmd
	StdErr  string
	ExecErr error
}

func (c CmdErr) Error() string {
	r := c.ExecErr.Error()
	if len(c.StdErr) > 0 {
		r += ": " + c.StdErr
	}
	return r
}

func (host *Host) Command(ctx context.Context, path string, args ...string) *Cmd {
	return &Cmd{
		Path: path,
		Args: args,

		Context: ctx,

		host: host,
	}
}

func (cmd *Cmd) Run() error {
	errbuf := &bytes.Buffer{}

	if cmd.Stderr == nil {
		cmd.Stderr = errbuf
	}

	if !cmd.host.SSH {
		c := exec.CommandContext(cmd.Context, cmd.Path, cmd.Args...)
		c.Stdin = cmd.Stdin
		c.Stdout = cmd.Stdout
		c.Stderr = cmd.Stderr
		if len(cmd.Env) > 0 {
			// Maybe just change this to copy everything?
			c.Env = append(c.Env, "PATH="+os.Getenv("PATH"))
			for _, e := range cmd.Env {
				c.Env = append(c.Env, e[0]+"="+e[1])
			}
		}
		if err := c.Run(); err != nil {
			return &CmdErr{Cmd: cmd, StdErr: strings.TrimSpace(errbuf.String()), ExecErr: err}
		}
		return nil
	}

	session, err := cmd.host.Run.Pool.Get(cmd.host.Host)
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdin = cmd.Stdin
	session.Stdout = cmd.Stdout
	session.Stderr = cmd.Stderr

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

	if cmd.host.Run.Verbose {
		fmt.Println("sshexec", cmd.host.Host, cmdline, cmd.Env)
	}

	err = session.Run(cmdline)

	if cmd.host.Run.Verbose {
		fmt.Println("sshexec", cmd.host.Host, cmdline, err)
	}

	if err != nil {
		return &CmdErr{Cmd: cmd, StdErr: strings.TrimSpace(errbuf.String()), ExecErr: err}
	}
	return nil
}
