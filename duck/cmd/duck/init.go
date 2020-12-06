package main

import (
	"os"
	"os/exec"
)

func initialize() error {
	if _, err := os.Stat("go.mod"); err != nil {
		cmd := exec.Command("go", "mod", "init")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}
