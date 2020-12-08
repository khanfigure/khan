package duck

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
)

type File struct {
	Path    string
	Content string
	User    User

	id int
}

func (f *File) String() string {
	return f.Path
}

func (f *File) setID(id int) {
	f.id = id
}
func (f *File) getID() int {
	return f.id
}
func (f *File) apply(r *run) error {
	buf, err := ioutil.ReadFile(f.Path)
	if err == nil && bytes.Compare(buf, []byte(f.Content)) == 0 {
		// no change
		return nil
	}
	if err != nil && iserrnotfound(err) {
		fmt.Printf("+ %s\n", f.Path)
	} else {
		reason := "content"
		if err != nil {
			reason = err.Error()
		}
		fmt.Printf("~ %s (%s)\n", f.Path, reason)
	}
	fh, err := os.Create(f.Path)
	if err != nil {
		return err
	}
	defer fh.Close()
	if _, err := fh.Write([]byte(f.Content)); err != nil {
		return err
	}
	if err := fh.Close(); err != nil {
		return err
	}
	return nil
}

func iserrnotfound(err error) bool {
	// TODO do this better
	v, ok := err.(*os.PathError)
	if ok && v != nil && v.Err == syscall.ENOENT {
		return true
	}
	return false
}
