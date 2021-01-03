package dry

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"syscall"
	"time"

	"github.com/desops/khan/rio/util"
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
	w.file.info.size = int64(len(w.file.content))
	w.file.info.modtime = time.Now()

	w.buf = nil
	w.err = nil
	return nil
}

func (host *Host) Create(fpath string) (io.WriteCloser, error) {
	host.fsmu.Lock()
	defer host.fsmu.Unlock()

	file := host.fs[fpath]
	if file != nil && file.info != nil {
		// TODO: Someday, be super cool and emulate a bunch of common permission errors.

		if file.info.isdir {
			return nil, &os.PathError{Op: "open", Path: fpath, Err: syscall.EISDIR}
		}
	}

	// simulate truncating whatever is there
	file = &File{
		info: &FileInfo{
			name:    path.Base(fpath),
			size:    0,
			mode:    0644,
			modtime: time.Now(),
			isdir:   false,
			uid:     host.uid,
			gid:     host.gid,
		},
	}
	host.fs[fpath] = file

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

		if file.info.isdir {
			return &os.PathError{Op: "rm", Path: fpath, Err: syscall.EISDIR}
		}
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
			info: &FileInfo{
				name:    f.Name(),
				size:    f.Size(),
				mode:    f.Mode(),
				modtime: f.ModTime(),
				isdir:   f.IsDir(),
			},
		}
		switch st := f.Sys().(type) {
		case *syscall.Stat_t:
			file.info.uid = st.Uid
			file.info.gid = st.Gid
		case *util.FileInfo:
			file.info.uid = st.Uid()
			file.info.gid = st.Gid()
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

	file.info.mode = mode
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
			info: &FileInfo{
				name:    f.Name(),
				size:    f.Size(),
				mode:    f.Mode(),
				modtime: f.ModTime(),
				isdir:   f.IsDir(),
			},
		}
		switch st := f.Sys().(type) {
		case *syscall.Stat_t:
			file.info.uid = st.Uid
			file.info.gid = st.Gid
		case *util.FileInfo:
			file.info.uid = st.Uid()
			file.info.gid = st.Gid()
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

	file.info.uid = uid
	file.info.gid = gid
	return nil
}
