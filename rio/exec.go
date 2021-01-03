package host

import (
	"context"
	"io"

	"github.com/keegancsmith/shell"
)

type Cmd struct {
	Path string
	Args []string
	Env  [][2]string
	Dir  string

	Shell bool // hack for when ssh-ing, try to start a shell instead of just executing a command (gives you working environment vars, etc)

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	Context context.Context

	// ReadOnly indicates that this command does not have side effects, and is safe to run in dry-run mode.
	ReadOnly bool
}

func (cmd *Cmd) String() string {
	s := "$ " + cmd.Path
	for _, arg := range cmd.Args {
		s += " " + shell.ReadableEscapeArg(arg)
	}
	return s
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

func Command(ctx context.Context, path string, args ...string) *Cmd {
	return &Cmd{
		Path:    path,
		Args:    args,
		Context: ctx,
	}
}

func ReadOnlyCommand(ctx context.Context, path string, args ...string) *Cmd {
	return &Cmd{
		Path:     path,
		Args:     args,
		Context:  ctx,
		ReadOnly: true,
	}
}
