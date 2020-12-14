package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

func build() error {
	// Enter a private space in /tmp so that we don't clutter the cwd with
	// generated intermediate crap. This space will persist: it is based on
	// the working directory absolute path. This way the go compiler can
	// optimize multiple runs.

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// safe, clean, slow way
	//	wd, err := ioutil.TempDir("", "duck")
	//	if err != nil {
	//		return err
	//	}

	// more fun way
	wd, err := stabletmpdir(cwd)
	if err != nil {
		return err
	}

	if err := sync(wd, cwd); err != nil {
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

	assetfs := false

	for _, match := range matches {
		base := filepath.Base(match)
		goname := base + ".go"
		if err := yaml2go(wd, match, wd+"/"+goname, &assetfs); err != nil {
			return err
		}
	}

	if _, err := os.Stat(wd + "/main.go"); err != nil {

		var assetfn = `func assetfn(path string) (io.Reader, error) {
	buf, err := Asset(path)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf), nil
}
`
		if !assetfs {
			assetfn = `func assetfn(path string) (io.Reader, error) {
	_ = bytes.NewReader
	return nil, os.ErrNotExist
}
`
		}

		if err := ioutil.WriteFile(wd+"/main.go", []byte(fmt.Sprintf(`package main
import (
	"fmt"
	"os"
	"io"
	"bytes"

	%s %#v
)
%s
func main() {
	if err := %s.Apply(assetfn); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
`, duckpkgalias, duckpkgname, assetfn, duckpkgalias)), 0644); err != nil {
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

	// Copy back out files that go often changes
	for _, p := range []string{"go.sum", "go.mod"} {
		if _, err := os.Stat(cwd + "/" + p); err == nil {
			err = compare(cwd+"/"+p, wd+"/"+p)
			if err == nil {
				continue
			}
			if err == errNotSame {
				if err := cp(cwd+"/"+p, wd+"/"+p); err != nil {
					return err
				}
			}
			return err
		}
	}

	return nil
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

func stabletmpdir(srcpath string) (string, error) {
	p, err := filepath.EvalSymlinks(srcpath)
	if err != nil {
		return "", err
	}

	h := fmt.Sprintf("%x", md5.Sum([]byte(p)))
	r := "/tmp/duck_" + h[:8]

	info, err := os.Stat(r)
	if err == nil && info.IsDir() {
		return r, nil
	}

	fmt.Println("mkdir", r)
	if err := os.Mkdir(r, 0700); err != nil {
		return "", err
	}
	return r, nil
}
