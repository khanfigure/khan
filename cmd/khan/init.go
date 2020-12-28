package main

import (
	"os"
	"os/exec"
	"path"
)

func initialize() error {
	if _, err := os.Stat("go.mod"); err != nil {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		base := path.Base(wd)

		cmd := exec.Command("go", "mod", "init", base)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	if _, err := os.Stat("go.sum"); err != nil {
		fh, err := os.Create("go.sum")
		if err != nil {
			return err
		}
		fh.Close()
	}

	return nil
}
