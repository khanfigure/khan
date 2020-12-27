package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var errNotSame = errors.New("File contents are different")

// sync is a recursive file copier that aggressively compare things in order to take no action if the
// copy would be a no-op. The idea is to let the go compiler cache things in our temporary folder, and
// to not modify mtime or anything else that we don't have to. It also deletes files in dest/ that are
// not in src/ without asking.
func sync(dest, src string) error {
	seen := map[string]struct{}{}

	copier := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasPrefix(path, src) {
			return fmt.Errorf("FileWalker src path substring fail: %#v %#v", path, src)
		}
		p := strings.TrimPrefix(path, src)
		seen[p] = struct{}{}

		//fmt.Println("copier path", p)

		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
		} else {
			if !(strings.HasSuffix(p, ".go") || info.Name() == "go.mod" || info.Name() == "go.sum") {
				return nil
			}
		}

		destp := dest + p

		destinfo, desterr := os.Stat(destp)

		if info.IsDir() {
			if desterr == nil && destinfo.IsDir() {
				return nil
			}
			//fmt.Println("mkdir", destp)
			return os.Mkdir(destp, 0700)
		}

		if desterr == nil && destinfo.Size() == info.Size() {
			err := compare(destp, path)
			if err == nil {
				//fmt.Println("files match", destp, path)
				// file contents match
				return nil
			}

			if err != errNotSame {
				// io error during compare
				return err
			}

			// files not same
			//fmt.Println("file contents changed:", destp, path)
			//fmt.Println("copying", path)
		}

		return cp(destp, path)
	}

	pruner := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasPrefix(path, dest) {
			return fmt.Errorf("FileWalker dest path substring fail: %#v %#v", path, dest)
		}
		p := strings.TrimPrefix(path, dest)
		//fmt.Println("pruner path", p)

		if _, ok := seen[p]; ok {
			return nil
		}

		// delete it

		//fmt.Println("deleting", path)

		if info.IsDir() {
			if err := os.RemoveAll(path); err != nil {
				return err
			}
			return filepath.SkipDir
		} else {
			return os.Remove(path)
		}

		return nil
	}

	if err := filepath.Walk(src, copier); err != nil {
		return err
	}

	if err := filepath.Walk(dest, pruner); err != nil {
		return err
	}

	return nil
}

func compare(p1, p2 string) error {
	var (
		buf1 [8192]byte
		buf2 [8192]byte
	)

	fh1, err := os.Open(p1)
	if err != nil {
		return err
	}
	defer fh1.Close()

	fh2, err := os.Open(p2)
	if err != nil {
		return err
	}
	defer fh2.Close()

	for {
		n1, err1 := fh1.Read(buf1[:])
		n2, err2 := fh2.Read(buf2[:])

		if err1 == io.EOF && err2 == io.EOF {
			// files are the same!
			return nil
		}
		if err1 == io.EOF || err2 == io.EOF {
			return errNotSame
		}
		if err1 != nil {
			return err1
		}
		if err2 != nil {
			return err2
		}

		// short read on n1
		for n1 < n2 {
			more, err := fh1.Read(buf1[n1:n2])
			if err == io.EOF {
				return errNotSame
			}
			if err != nil {
				return err
			}
			n1 += more
		}
		// short read on n2
		for n2 < n1 {
			more, err := fh2.Read(buf2[n2:n1])
			if err == io.EOF {
				return errNotSame
			}
			if err != nil {
				return err
			}
			n2 += more
		}
		if n1 != n2 {
			// should never happen unless i have a logic bug above
			return fmt.Errorf("file compare reads out of sync: %d != %d", n1, n2)
		}

		if bytes.Compare(buf1[:n1], buf2[:n2]) != 0 {
			return errNotSame
		}
	}
}

func cp(dest, src string) error {
	//fmt.Println("copying", src, "→", dest)
	destfh, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer destfh.Close()

	srcfh, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcfh.Close()

	if _, err := io.Copy(destfh, srcfh); err != nil {
		return err
	}

	return destfh.Close()
}
