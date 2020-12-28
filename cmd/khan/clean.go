package main

import (
	"os"
	"os/exec"
	"strings"
)

func clean() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	wd, err := stabletmpdir(cwd)
	if err != nil {
		return err
	}

	buf, err := exec.Command("go", "list", "-m").Output()
	if err != nil {
		return err
	}
	outfile := strings.TrimSpace(string(buf))

	if _, err := os.Stat(outfile); err == nil {
		if err := os.Remove(outfile); err != nil {
			return err
		}
	}

	return os.RemoveAll(wd)
}
