package remote

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"syscall"

	"github.com/keegancsmith/shell"
)

type Reader struct {
	reader *io.PipeReader

	closemu  sync.Mutex
	closed   bool
	closeerr error
	procerr  chan error
}

func (r *Reader) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}
func (r *Reader) Close() error {
	r.closemu.Lock()
	defer r.closemu.Unlock()

	if r.closed {
		return r.closeerr
	}

	r.closed = true
	cerr := r.reader.Close()
	r.closeerr = <-r.procerr
	if r.closeerr == nil && cerr != nil {
		r.closeerr = cerr
	}

	return r.closeerr
}

func (host *Host) Open(path string) (io.ReadCloser, error) {
	reader := &Reader{
		procerr: make(chan error),
	}

	session, err := host.pool.Get(host.connect)
	if err != nil {
		return nil, err
	}

	r, w := io.Pipe()

	errbuf := &bytes.Buffer{}

	session.Stdout = w
	session.Stderr = errbuf
	reader.reader = r

	cmdline := "cat " + shell.ReadableEscapeArg(path)

	if err := session.Start(cmdline); err != nil {
		w.Close()
		r.Close()
		session.Put()
		return nil, err
	}

	go func() {
		err := session.Wait()
		e := strings.TrimSpace(errbuf.String())

		if err != nil {
			if strings.HasPrefix(e, "cat: ") && strings.HasSuffix(e, "No such file or directory") {
				// emulate os.Open
				err = &os.PathError{
					Op:   "open",
					Path: path,
					Err:  syscall.ENOENT,
				}
			} else {
				// Bundle up stderr and hope it's useful
				err = fmt.Errorf("Command %#v on host %#v: %w: %s",
					cmdline, host.connect, err, e)
			}
		}

		// This will let blocked reads finish
		w.CloseWithError(err)

		reader.procerr <- err
		close(reader.procerr)
		session.Put()
	}()

	return reader, nil
}

func (host *Host) ReadFile(fpath string) ([]byte, error) {
	buf := &bytes.Buffer{}
	fh, err := host.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	_, err = io.Copy(buf, fh)
	if err != nil {
		return nil, err
	}
	err = fh.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
