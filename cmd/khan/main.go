package main

import (
	"fmt"
	"os"
)

var cmds = map[string]func() error{
	"build": build,
	"init":  initialize,
	//"go": gocmd, // At first I thought I'd need this but now I can't find a use for it
	"clean": clean,
}

func main() {
	if len(os.Args) < 2 {
		return
	}
	cmd := os.Args[1]
	fn, ok := cmds[cmd]
	if !ok {
		if cmd != "-h" && cmd != "--help" {
			fmt.Fprintf(os.Stderr, "Unknown command %#v\n", cmd)
		}
		fmt.Fprintf(os.Stderr, `
Usage:

    %s <command>

Commands:

`, os.Args[0])
		for cmd := range cmds {
			fmt.Fprintf(os.Stderr, "    %s\n", cmd)
		}

		os.Exit(1)
	}

	if err := fn(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
