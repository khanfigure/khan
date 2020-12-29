package khan

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

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

		if host.Run.Dry {
			host.VirtMu.RLock()
			fi, hit := host.Virt.Files[f.Path]
			host.VirtMu.RUnlock()
			if hit && fi == nil {
				return itemUnchanged, nil
			}
		}

		_, err := host.Stat(f.Path)
		if err != nil && iserrnotfound(err) {
			return itemUnchanged, nil
		}
		if err != nil {
			// You know... I think this isn't safe and we should
			// return the error. It's probably a permission error.
			// I keep going back and forth on this.
			return 0, err
		}

		if host.Run.Verbose && err != nil {
			//fmt.Fprintln(os.Stderr, "stat", f.Path, "error:", err)
		}

		if !host.Run.Dry {
			err := host.Remove(f.Path)
			if err != nil {
				return 0, err
			}
		}

		host.VirtMu.Lock()
		host.Virt.Files[f.Path] = nil
		delete(host.Virt.Content, f.Path)
		host.VirtMu.Unlock()

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

	var (
		buf    []byte
		err    error
		cached bool
	)

	if host.Run.Dry {
		host.VirtMu.RLock()
		cfi, fihit := host.Virt.Files[f.Path]
		ccontent := host.Virt.Content[f.Path]
		host.VirtMu.RUnlock()
		if fihit {
			cached = true
			if cfi == nil {
				err = &os.PathError{
					Op:   "read",
					Path: f.Path,
					Err:  syscall.ENOENT,
				}
			} else {
				buf = []byte(ccontent)
			}
		}
	}

	if !cached {
		buf, err = host.ReadFile(f.Path)
	}

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

	if host.Run.Dry {
		host.VirtMu.Lock()
		host.Virt.Files[f.Path] = &FileInfo{
			name:    f.Path,
			size:    int64(len(content)),
			mode:    0644, // TODO
			modtime: time.Now(),
			isdir:   false,
			// TODO uid and gid
		}
		host.Virt.Content[f.Path] = content
		host.VirtMu.Unlock()
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

func (f *File) applyperms(host *Host) (itemStatus, error) {

	// need this data to resolve uid->name and gid->name of files
	if err := host.getUserGroups(false); err != nil {
		return 0, err
	}

	mode := f.Mode
	if mode == 0 {
		mode = 0644
	}

	v := host.Virt

	usr := f.User
	if usr == "" {
		if host.SSH {
			usr = host.Run.User
		} else {
			usr = os.Getenv("USER")
		}
	}

	if usr == "" {
		return 0, fmt.Errorf("Cannot determine user for managed file %v", f)
	}

	grp := f.Group
	if grp == "" {
		grp = usr
		// Actually, default to the login group of the user if we can
		host.VirtMu.Lock()
		u := v.cacheUsers[usr]
		if host.Run.Dry {
			cu, hit := v.Users[usr]
			if hit {
				u = cu
			}
		}
		if u != nil {
			grp = u.Group
		}
		host.VirtMu.Unlock()
	}

	if grp == "" {
		return 0, fmt.Errorf("Cannot determine group for managed file %v", f)
	}

	var (
		uid     uint32
		gid     uint32
		wantuid uint32
		wantgid uint32
	)

	host.VirtMu.RLock()
	wantuser := v.cacheUsers[usr]
	wantgroup := v.cacheGroups[grp]
	if host.Run.Dry {
		cu, uok := v.Users[usr]
		if uok {
			wantuser = cu
		}
		cg, gok := v.Groups[grp]
		if gok {
			wantgroup = cg
		}
	}
	if wantuser != nil {
		wantuid = wantuser.Uid
	}
	if wantgroup != nil {
		wantgid = wantgroup.Gid
	}
	cachefi, hit := v.Files[f.Path]
	host.VirtMu.RUnlock()

	if wantuser == nil {
		return 0, fmt.Errorf("Unknown user %#v", f.User)
	}
	if wantgroup == nil {
		return 0, fmt.Errorf("Unknown group %#v", f.Group)
	}

	var fi os.FileInfo

	if host.Run.Dry && hit {
		fi = cachefi
	} else {
		ofi, err := host.Stat(f.Path)
		if err != nil {
			return 0, err
		}
		fi = ofi
	}

	switch st := fi.Sys().(type) {
	case *syscall.Stat_t:
		uid = st.Uid
		gid = st.Gid
	case *FileInfo:
		uid = st.uid
		gid = st.gid
	default:
		return 0, fmt.Errorf("Unhandled system stat type %T", fi.Sys())
	}

	status := itemUnchanged

	newfi := &FileInfo{
		name:    fi.Name(),
		size:    fi.Size(),
		modtime: fi.ModTime(),
		isdir:   fi.IsDir(),
		uid:     wantuid,
		gid:     wantgid,
		mode:    mode,
	}

	if wantuid != uid || wantgid != gid {
		fmt.Printf("wantuid %d wantgid %d uid %d gid %d\n", wantuid, wantgid, uid, gid)
		status = itemModified

		if !host.Run.Dry {
			if err := host.Chown(f.Path, wantuid, wantgid); err != nil {
				return 0, err
			}
		}

	}

	if fi.Mode()&s_justmode != mode {
		fmt.Printf("current: %o , masked %o , want: %o\n", uint32(fi.Mode()), uint32(fi.Mode())&s_justmode, mode)
		status = itemModified

		if !host.Run.Dry {
			if err := host.Chmod(f.Path, mode); err != nil {
				return 0, err
			}
		}
	}

	if status == itemModified {
		host.VirtMu.Lock()
		v.Files[f.Path] = newfi
		host.VirtMu.Unlock()
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
