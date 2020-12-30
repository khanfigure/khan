package dry

import (
	"bytes"
	"io"
	"os"
	"path"
	"syscall"
	"time"
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
