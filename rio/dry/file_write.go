package dry

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"syscall"
	"time"

	"khan.rip/rio/util"
)

type Writer struct {
	file *File
	buf  *bytes.Buffer
	host *Host
	err  error
}

func (w *Writer) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}
func (w *Writer) Close() error {
	w.host.fsmu.Lock()
	defer w.host.fsmu.Unlock()
	if w.buf == nil {
		// already closed
		return w.err
	}

	w.file.content = w.buf.Bytes()
	w.file.info.Fsize = int64(len(w.file.content))
	w.file.info.Fmodtime = time.Now()

	w.buf = nil
	w.err = nil
	return nil
}

func (host *Host) Create(fpath string) (io.WriteCloser, error) {
	fmt.Println(host, ">", fpath)

	host.fsmu.Lock()
	defer host.fsmu.Unlock()

	file := host.fs[fpath]
	if file != nil && file.info != nil {
		// TODO: Someday, be super cool and emulate a bunch of common permission errors.

		if file.info.Fisdir {
			return nil, &os.PathError{Op: "open", Path: fpath, Err: syscall.EISDIR}
		}
	}

	// simulate truncating whatever is there
	file = &File{
		info: &util.FileInfo{
			Fname:    path.Base(fpath),
			Fsize:    0,
			Fmode:    0644,
			Fmodtime: time.Now(),
			Fisdir:   false,
			Fuid:     host.uid,
			Fgid:     host.gid,
		},
	}
	host.fs[fpath] = file

	//host.debug()

	writer := &Writer{
		file: file,
		host: host,
		buf:  &bytes.Buffer{},
	}
	return writer, nil
}

func (host *Host) Remove(fpath string) error {
	host.fsmu.Lock()
	defer host.fsmu.Unlock()

	file := host.fs[fpath]
	if file != nil && file.info != nil {
		// TODO: Someday, be super cool and emulate a bunch of common permission errors.

		if file.info.Fisdir {
			return &os.PathError{Op: "rm", Path: fpath, Err: syscall.EISDIR}
		}
	}

	if err := util.Remove(host, fpath); err != nil {
		return err
	}

	host.fs[fpath] = &File{}
	return nil
}

func (host *Host) Chmod(fpath string, mode os.FileMode) error {
	host.fsmu.Lock()
	defer host.fsmu.Unlock()

	file := host.fs[fpath]
	if file == nil && host.cascade != nil {
		f, err := host.cascade.Stat(fpath)
		if err != nil {
			return err
		}
		file = &File{
			info: &util.FileInfo{
				Fname:    f.Name(),
				Fsize:    f.Size(),
				Fmode:    f.Mode(),
				Fmodtime: f.ModTime(),
				Fisdir:   f.IsDir(),
			},
		}
		switch st := f.Sys().(type) {
		case *syscall.Stat_t:
			file.info.Fuid = st.Uid
			file.info.Fgid = st.Gid
		case *util.FileInfo:
			file.info.Fuid = st.Uid()
			file.info.Fgid = st.Gid()
		default:
			fmt.Errorf("Unhandled stat type %T", f.Sys())
		}
		host.fs[fpath] = file
	}
	if file == nil || file.info == nil {
		return &os.PathError{Op: "chmod", Path: fpath, Err: syscall.ENOENT}
	}

	if err := util.Chmod(host, fpath, mode); err != nil {
		return err
	}

	file.info.Fmode = mode
	return nil
}

func (host *Host) Chown(fpath string, uid uint32, gid uint32) error {
	host.fsmu.Lock()
	defer host.fsmu.Unlock()

	file := host.fs[fpath]
	if file == nil && host.cascade != nil {
		f, err := host.cascade.Stat(fpath)
		if err != nil {
			return err
		}
		file = &File{
			info: &util.FileInfo{
				Fname:    f.Name(),
				Fsize:    f.Size(),
				Fmode:    f.Mode(),
				Fmodtime: f.ModTime(),
				Fisdir:   f.IsDir(),
			},
		}
		switch st := f.Sys().(type) {
		case *syscall.Stat_t:
			file.info.Fuid = st.Uid
			file.info.Fgid = st.Gid
		case *util.FileInfo:
			file.info.Fuid = st.Uid()
			file.info.Fgid = st.Gid()
		default:
			fmt.Errorf("Unhandled stat type %T", f.Sys())
		}
		host.fs[fpath] = file
	}
	if file == nil || file.info == nil {
		return &os.PathError{Op: "chown", Path: fpath, Err: syscall.ENOENT}
	}

	if err := util.Chown(host, fpath, uid, gid); err != nil {
		return err
	}

	file.info.Fuid = uid
	file.info.Fgid = gid
	return nil
}
