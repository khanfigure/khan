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

func (host *Host) Rename(fpath, newpath string) error {
	host.fsmu.Lock()
	defer host.fsmu.Unlock()

	sa, erra := host.stat(fpath)
	sb, errb := host.stat(newpath)

	if errb == nil && sb.IsDir() {
		return fmt.Errorf("mv: cannot overwrite directory %#v", newpath)
	}

	if erra == nil && sa.IsDir() {
		if errb == nil && !sb.IsDir() {
			return fmt.Errorf("mv: cannot overwrite non-directory %#v with directory %#v", newpath, fpath)
		}
	}

	file := host.fs[fpath]

	// we need to read the existing contents in order to correctly model the move, otherwise a dry run
	// rename followed by a read would not return the correct contents. Maybe in the future, this could
	// be replaced by a sort of virtual symlink to the cascade filesystem's path?
	if file == nil && host.cascade != nil {
		buf, err := host.cascade.ReadFile(fpath)
		if err != nil {
			return err
		}

		fi, err := util.ConvertStat(sa)
		if err != nil {
			return err
		}

		file = &File{
			info:    fi,
			content: buf,
		}
	}

	if file == nil || file.info == nil {
		return &os.PathError{Op: "mv", Path: fpath, Err: syscall.ENOENT}
	}

	if err := util.Rename(host, fpath, newpath); err != nil {
		return err
	}

	host.fs[fpath] = &File{}
	host.fs[newpath] = file

	file.info.Fname = path.Base(newpath)

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

		fi, err := util.ConvertStat(f)
		if err != nil {
			return err
		}

		file = &File{
			info: fi,
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
		info, err := util.ConvertStat(f)
		if err != nil {
			return err
		}

		file = &File{
			info: info,
		}
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
