package duck

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"

	//"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/pmezard/go-difflib/difflib"
)

type File struct {
	Path string
	User User

	Content  string
	Template string

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

func (f *File) Validate() error {
	if f.Path == "" {
		return errors.New("File path is required")
	}
	if f.Content != "" && f.Template != "" {
		return errors.New("File content and template cannot both be specified")
	}
	return nil
}

func (f *File) StaticFiles() []string {
	if f.Template != "" {
		return []string{f.Template}
	}
	return nil
}

func (f *File) apply(r *run) error {
	r.addStat("files")

	content := f.Content

	if f.Template != "" {
		var err error
		content, err = executeTemplate(r, f.Template)
		if err != nil {
			return err
		}
	}

	buf, err := ioutil.ReadFile(f.Path)
	if err == nil && bytes.Compare(buf, []byte(content)) == 0 {
		if r.verbose {
			fmt.Printf("  %s up to date\n", f.Path)
		}
		r.addStat("files up to date")
		return nil
	}
	if err != nil && iserrnotfound(err) {
		fmt.Printf("+ %s\n", f.Path)
		r.addStat("files new")
	} else {
		reason := "content"
		if f.Template != "" {
			reason += " from template"
		}
		if err != nil {
			reason = err.Error()
		}
		fmt.Printf("~ %s (%s)\n", f.Path, reason)
		r.addStat("files content changed")
	}

	if r.diff {
		// This is cute but actually ugly.
		//dmp := diffmatchpatch.New()
		//diffs := dmp.DiffMain(string(buf), content, true)
		//fmt.Println(dmp.DiffPrettyText(diffs))
		diff := difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(buf)),
			B:        difflib.SplitLines(content),
			FromFile: f.Path,
			ToFile:   f.Path,
			Context:  3,
		}
		difftxt, err := difflib.GetUnifiedDiffString(diff)
		if err != nil {
			return err
		}
		fmt.Print(difftxt)
	}

	if r.dry {
		return nil
	}

	fh, err := os.Create(f.Path)
	if err != nil {
		return err
	}
	defer fh.Close()
	if _, err := fh.Write([]byte(content)); err != nil {
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
