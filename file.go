package khan

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"syscall"

	//"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/pmezard/go-difflib/difflib"
)

type File struct {
	Path string
	User User

	Mode os.FileMode

	Content  string
	Template string

	Delete bool

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
	if f.Delete && (f.Content != "" || f.Template != "") {
		return fmt.Errorf("File remove: true conflicts with content or template")
	}
	if f.Content != "" && f.Template != "" {
		//return errors.New("File content and template cannot both be specified")
		return fmt.Errorf("File content and template cannot both be specified (%#v %#v)", f.Content, f.Template)
	}
	return nil
}

func (f *File) StaticFiles() []string {
	if f.Template != "" {
		return []string{f.Template}
	}
	return nil
}

func (f *File) apply(r *run) (itemStatus, error) {
	status := itemModified

	if f.Delete {
		_, err := r.rioconfig.Stat(f.Path)
		if err != nil && iserrnotfound(err) {
			return itemUnchanged, nil
		}
		if r.verbose && err != nil {
			fmt.Fprintln(os.Stderr, "stat", f.Path, "error:", err)
		}

		status = itemDeleted

		if r.dry {
			return status, nil
		}
		return status, r.rioconfig.Remove(f.Path)
	}

	content := f.Content
	if f.Template != "" {
		var err error
		content, err = executeTemplate(r, f.Template)
		if err != nil {
			return 0, err
		}
	}

	buf, err := r.rioconfig.ReadFile(f.Path)
	if err == nil && bytes.Compare(buf, []byte(content)) == 0 {
		return itemUnchanged, nil
	}
	if err != nil {
		if iserrnotfound(err) {
			status = itemCreated
		} else {
			// This seemed risky.
			// If the file could not be read, don't assume
			// we should continue with writing to it.
			//		status = itemModified
			//		if err != nil {
			//			if r.verbose {
			//				fmt.Printf("Error reading %#v: %v\n", f.Path, err)
			//			}
			//		}

			// Instead let's return the read error.
			return 0, err
		}
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
			return 0, err
		}
		fmt.Print(difftxt)
	}

	if r.dry {
		return status, nil
	}

	fh, err := r.rioconfig.Create(f.Path)
	if err != nil {
		return 0, err
	}
	defer fh.Close()
	if _, err := fh.Write([]byte(content)); err != nil {
		return 0, err
	}
	if err := fh.Close(); err != nil {
		return 0, err
	}
	return status, nil
}

func iserrnotfound(err error) bool {
	// TODO do this better
	v, ok := err.(*os.PathError)
	if ok && v != nil && v.Err == syscall.ENOENT {
		return true
	}
	return false
}
