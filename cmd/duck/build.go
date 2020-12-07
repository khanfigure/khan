package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func build() error {
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
		basename := strings.TrimSuffix(strings.TrimSuffix(match, ".yaml"), ".yml")
		goname := basename + "_fromyaml.go"
		if err := yaml2go(match, goname); err != nil {
			return err
		}
	}

	if _, err := os.Stat("main.go"); err != nil {
		if err := ioutil.WriteFile("mainmain_fromyaml.go", []byte(fmt.Sprintf(`package main
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

	cmd := exec.Command("go", "build")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
