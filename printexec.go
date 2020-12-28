package khan

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/keegancsmith/shell"
)

func printExec(host *Host, c string, args ...string) error {
	return printExecStdin(host, nil, c, args...)
}

func printExecStdin(host *Host, stdin io.Reader, c string, args ...string) error {
	r := host.Run

	if r.Verbose {
		fmt.Print(shell.ReadableEscapeArg(c))
		for _, a := range args {
			fmt.Print(" " + shell.ReadableEscapeArg(a))
		}
		fmt.Println()
	}
	if r.Dry {
		return nil
	}
	cmd := host.Command(context.Background(), c, args...)
	cmd.Stdin = stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
