package khan

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"strings"
	"syscall"

	"github.com/desops/khan/rio/util"

	"github.com/pmezard/go-difflib/difflib"
)

type File struct {
	Path string `khan:"path,shortkey"`

	User  string
	Group string
	Mode  os.FileMode

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

func (f *File) After() []string {
	if f.Delete {
		return nil
	}
	var afters []string
	if f.Local != "" {
		afters = append(afters, "path:"+f.Local)
	}

	if f.User != "" {
		afters = append(afters, "user:"+f.User)
	}
	if f.Group != "" {
		afters = append(afters, "group:"+f.Group)
	}
	return afters
}
func (f *File) Before() []string {
	return nil
}
func (f *File) Provides() []string {
	return []string{"path:" + f.Path}
}

func (f *File) Apply(host *Host) (itemStatus, error) {
	status := itemModified

	if f.Delete {
		_, err := host.rh.Stat(f.Path)
		if err != nil && iserrnotfound(err) {
			return itemUnchanged, nil
		}
		if err != nil {
			return 0, err
		}
		if err := host.rh.Remove(f.Path); err != nil {
			return 0, err
		}
		return itemDeleted, nil
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
			srcbuf, err := host.rh.ReadFile(f.Local)
			if err != nil {
				return 0, err
			}
			content = string(srcbuf)
		}
		// else: assume Content is the content. (Blank means a blank file.)
	} else {
		return 0, fmt.Errorf("Unknown template engine %#v", engine)
	}

	var (
		buf []byte
		err error
	)

	buf, err = host.rh.ReadFile(f.Path)

	if err == nil && bytes.Compare(buf, []byte(content)) == 0 {
		pstatus, err := f.applyperms(host)
		if err != nil {
			return 0, err
		}
		return pstatus, nil
	}
	if err != nil {
		if iserrnotfound(err) {
			status = itemCreated
		} else {
			return 0, err
		}
	}

	if host.Run.Diff {
		// This is cute but actually ugly.
		// import "github.com/sergi/go-diff/diffmatchpatch"
		//dmp := diffmatchpatch.New()
		//diffs := dmp.DiffMain(string(buf), content, true)
		//fmt.Println(dmp.DiffPrettyText(diffs))

		// this seems nicer
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

	fh, err := host.rh.Create(f.Path)
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

func (f *File) applyperms(host *Host) (itemStatus, error) {
	mode := f.Mode
	if mode == 0 {
		mode = 0644
	}

	ustr := f.User
	if ustr == "" {
		at := strings.IndexByte(host.Host, '@')
		if host.SSH && at > -1 {
			ustr = host.Host[:at]
		} else {
			osu, err := user.Current()
			if err != nil {
				return 0, err
			}
			ustr = osu.Username
		}
	}
	if ustr == "" {
		return 0, fmt.Errorf("Cannot determine user for managed file %v", f)
	}

	user, err := host.rh.User(ustr)
	if err != nil {
		return 0, err
	}

	gstr := f.Group
	if gstr == "" {
		gstr = user.Group
	}

	if gstr == "" {
		return 0, fmt.Errorf("Cannot determine group for managed file %v", f)
	}

	group, err := host.rh.Group(gstr)
	if err != nil {
		return 0, err
	}

	var (
		uid     uint32
		gid     uint32
		wantuid uint32
		wantgid uint32
	)

	if user == nil {
		return 0, fmt.Errorf("Unknown user %#v", ustr)
	}
	if group == nil {
		return 0, fmt.Errorf("Unknown group %#v", gstr)
	}

	fi, err := host.rh.Stat(f.Path)
	if err != nil {
		return 0, err
	}

	switch st := fi.Sys().(type) {
	case *syscall.Stat_t:
		uid = st.Uid
		gid = st.Gid
	case *util.FileInfo:
		uid = st.Uid()
		gid = st.Gid()
	default:
		return 0, fmt.Errorf("Unhandled system stat type %T", fi.Sys())
	}

	status := itemUnchanged

	if wantuid != uid || wantgid != gid {
		fmt.Printf("wantuid %d wantgid %d uid %d gid %d\n", wantuid, wantgid, uid, gid)
		status = itemModified

		if err := host.rh.Chown(f.Path, wantuid, wantgid); err != nil {
			return 0, err
		}
	}

	if fi.Mode()&util.S_justmode != mode {
		fmt.Printf("current: %o , masked %o , want: %o\n", uint32(fi.Mode()), uint32(fi.Mode())&util.S_justmode, mode)
		status = itemModified

		if err := host.rh.Chmod(f.Path, mode); err != nil {
			return 0, err
		}
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
