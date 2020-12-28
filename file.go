package khan

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"

	//"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/pmezard/go-difflib/difflib"
)

type File struct {
	Path string `khan:"path,shortkey"`
	User User

	Mode os.FileMode

	// Content specifies a static string for the content of the file.
	Content string

	// Src is a path on the configurer for the source of the file.
	// This will be bundled into your khan build output.
	Src string `khan:"src,shortvalue"`

	// Local is a path on the configuree for the source of the file
	Local string

	// Template execution mode. Leave blank for no templating. Special
	// value "1" is the same as the default templating engine "pongo2",
	// a jinja2 style template engine. (See https://github.com/flosch/pongo2)
	Template string

	Delete bool

	id int
}

func (f *File) String() string {
	return f.Path
}

func (f *File) SetID(id int) {
	f.id = id
}
func (f *File) ID() int {
	return f.id
}
func (f *File) Clone() Item {
	r := *f
	r.id = 0
	return &r
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
	if f.Src != "" {
		return []string{f.Src}
	}
	return nil
}

func (f *File) Needs() []string {
	if f.Local != "" {
		return []string{"path:" + f.Local}
	}
	return nil
}
func (f *File) Provides() []string {
	return []string{"path:" + f.Path}
}

func (f *File) Apply(host *Host) (itemStatus, error) {
	status := itemModified

	if f.Delete {
		_, err := host.Stat(f.Path)
		if err != nil && iserrnotfound(err) {
			return itemUnchanged, nil
		}
		if host.Run.Verbose && err != nil {
			fmt.Fprintln(os.Stderr, "stat", f.Path, "error:", err)
		}

		status = itemDeleted

		if host.Run.Dry {
			return status, nil
		}
		return status, host.Remove(f.Path)
	}

	content := f.Content

	engine := f.Template
	if engine == "1" || engine == "true" || engine == "yes" || engine == "pongo" {
		engine = "pongo2"
	}

	if engine == "pongo2" {
		if f.Src != "" {
			var err error
			if content, err = executePackedTemplateFile(host, f.Src); err != nil {
				return 0, err
			}
		} else if f.Local != "" {
			return 0, fmt.Errorf("FIXME template local mode not supported yet. (security considerations?)")
		} else {
			var err error
			if content, err = executePackedTemplateString(host, f.Content); err != nil {
				return 0, err
			}
		}
	} else if engine == "" {
		// raw file mode
		if f.Src != "" {
			fh, err := host.Run.assetfn(f.Src)
			if err != nil {
				return 0, err
			}
			defer fh.Close()
			buf := &bytes.Buffer{}
			if _, err := io.Copy(buf, fh); err != nil {
				return 0, err
			}
			content = buf.String()
		} else if f.Local != "" {
			// copy from another path on managed host
			srcbuf, err := host.ReadFile(f.Local)
			if err != nil {
				return 0, err
			}
			content = string(srcbuf)
		}
		// else: assume Content is the content. (Blank means a blank file.)
	} else {
		return 0, fmt.Errorf("Unknown template engine %#v", engine)
	}

	buf, err := host.ReadFile(f.Path)
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
			//			if host.Run.Verbose {
			//				fmt.Printf("Error reading %#v: %v\n", f.Path, err)
			//			}
			//		}

			// Instead let's return the read error.
			return 0, err
		}
	}

	if host.Run.Diff {
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

	if host.Run.Dry {
		return status, nil
	}

	fh, err := host.Create(f.Path)
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
