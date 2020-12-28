package main

import (
	"os"
	"os/exec"
)

func gocmd() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	wd, err := stabletmpdir(cwd)
	if err != nil {
		return err
	}

	args := os.Args[2:]

	cmd := exec.Command("go", args...)
	cmd.Dir = wd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
