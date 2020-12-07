package main

import (
	"path"
	"os"
	"os/exec"
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

	return nil
}
