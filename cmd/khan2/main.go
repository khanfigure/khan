package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"khan.rip/keval"
	"khan.rip/ksyn"
	//"github.com/mgutz/ansi"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	files, err := filepath.Glob("*.k")
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return errors.New("No .k files found, nothing to do")
	}

	machine := keval.NewMachine()

	for _, fpath := range files {
		ktree, err := ksyn.ParseFile(fpath)
		if err != nil {
			dump(err)
			return err
		}

		fmt.Println(ksyn.ReprUnicode(ktree))

		v, err := machine.Eval(ktree)
		if err != nil {
			dump(err)
			return err
		}
		printv(v.V, "")
	}

	return nil
}

func dump(err error) {
	var ne ksyn.ErrorFromNode
	if errors.As(err, &ne) {
		fmt.Fprintln(os.Stderr, ksyn.NodeSourceFragment(ne.Node))
	}

	var e ksyn.Error
	if errors.As(err, &e) {
		fmt.Fprintln(os.Stderr, ksyn.SourceFragment(e.Pos))
	}
}

func printv(vi interface{}, indent string) {
	if vi == nil {
		return
	}
	switch v := vi.(type) {
	case string:
		fmt.Printf("%s%#v\n", indent, v)
	case int:
		fmt.Printf("%s%d\n", indent, v)
	case map[string]interface{}:
		fmt.Printf("%s{\n", indent)
		for k, vv := range v {
			fmt.Printf("%s%#v:\n", indent+"   ", k)
			printv(vv, indent+"   ")
		}
		fmt.Printf("%s}\n", indent)
	case []interface{}:
		fmt.Printf("%s[\n", indent)
		for _, vv := range v {
			printv(vv, indent+"   ")
		}
		fmt.Printf("%s]\n", indent)
	}
	fmt.Printf("%s[%T]\n", indent, vi)
}
