package keval

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"khan.rip/ksyn"
)

func builtin_include(m *Machine, fpath string) {
	if strings.HasSuffix(fpath, "/") {
		files, err := filepath.Glob(fpath + "*.k")
		if err != nil {
			panic(err)
		}
		if len(files) == 0 {
			panic("No .k files found, nothing to do")
		}
		for _, f := range files {
			if err := include(m, f); err != nil {
				dumperr(err)
				panic(err)
			}
		}
	} else {
		if err := include(m, fpath); err != nil {
			dumperr(err)
			panic(err)
		}
	}
}

func include(m *Machine, fpath string) error {

	fmt.Println("INCLUDING", fpath)

	ktree, err := ksyn.ParseFile(fpath)
	if err != nil {
		return err
	}

	_, err = m.Eval(ktree)
	if err != nil {
		return err
	}

	return nil
}

func dumperr(err error) {
	var ne ksyn.ErrorFromNode
	if errors.As(err, &ne) {
		fmt.Fprintln(os.Stderr, ksyn.NodeSourceFragment(ne.Node))
	}

	var e ksyn.Error
	if errors.As(err, &e) {
		fmt.Fprintln(os.Stderr, ksyn.SourceFragment(e.Pos))
	}
}
