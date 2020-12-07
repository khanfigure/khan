package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

func build() error {

	// enter a private space in /tmp so that we don't clutter the cwd with
	// generated intermediate crap

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	wd, err := ioutil.TempDir("", "duck")
	if err != nil {
		return err
	}

	if err := copyglobs(wd, "*.go", "go.mod", "go.sum"); err != nil {
		return err
	}

	matches, err := filepath.Glob("*.yaml")
	if err != nil {
		return err
	}
	matches2, err := filepath.Glob("*.yml")
	if err != nil {
		return err
	}
	matches = append(matches, matches2...)
	sort.Strings(matches)

	for _, match := range matches {
		base := filepath.Base(match)
		goname := base + ".go"
		if err := yaml2go(match, wd+"/"+goname); err != nil {
			return err
		}
	}

	if _, err := os.Stat(wd + "/main.go"); err != nil {
		if err := ioutil.WriteFile(wd+"/main.go", []byte(fmt.Sprintf(`package main
import (
	"fmt"
	"os"

	%s %#v
)

func main() {
	if err := %s.Apply(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
`, duckpkgalias, duckpkgname, duckpkgalias)), 0644); err != nil {
			return err
		}
	}

	cmd := exec.Command("go", "build", "-o", cwd+"/duck")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = wd
	if err := cmd.Run(); err != nil {
		return err
	}

	// not in defer on purpose: I don't want to clean up the build folder
	// when there is an error for now, so I can debug.
	return os.RemoveAll(wd)
}

func copyglobs(dest string, globs ...string) error {
	for _, g := range globs {
		matches, err := filepath.Glob(g)
		if err != nil {
			return err
		}
		for _, match := range matches {
			base := filepath.Base(match)
			//dir := filepath.Dir(match)
			to := dest + "/" + base

			destfh, err := os.Create(to)
			if err != nil {
				return err
			}
			defer destfh.Close()
			srcfh, err := os.Open(match)
			if err != nil {
				return err
			}
			defer srcfh.Close()
			if _, err := io.Copy(destfh, srcfh); err != nil {
				return err
			}
			if err := destfh.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}
