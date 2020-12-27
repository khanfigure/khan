package khan

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/keegancsmith/shell"
)

func printExec(r *Run, c string, args ...string) error {
	return printExecStdin(r, nil, c, args...)
}

func printExecStdin(r *Run, stdin io.Reader, c string, args ...string) error {
	if r.verbose {
		fmt.Print(shell.ReadableEscapeArg(c))
		for _, a := range args {
			fmt.Print(" " + shell.ReadableEscapeArg(a))
		}
		fmt.Println()
	}
	if r.dry {
		return nil
	}
	cmd := r.rioconfig.Command(context.Background(), c, args...)
	cmd.Stdin = stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
