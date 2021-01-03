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

type Writer struct {
	writer *io.PipeWriter

	closemu  sync.Mutex
	closed   bool
	closeerr error
	procerr  chan error
}

func (w *Writer) Write(p []byte) (int, error) {
	return w.writer.Write(p)
}
func (w *Writer) Close() error {
	w.closemu.Lock()
	defer w.closemu.Unlock()

	if w.closed {
		return w.closeerr
	}

	w.closed = true
	cerr := w.writer.Close()
	w.closeerr = <-w.procerr
	if w.closeerr == nil && cerr != nil {
		w.closeerr = cerr
	}
	return w.closeerr
}

func (host *Host) Create(path string) (io.WriteCloser, error) {
	session, err := host.pool.Get(host.connect)
	if err != nil {
		return nil, err
	}

	r, w := io.Pipe()
	errbuf := &bytes.Buffer{}

	session.Stdin = r
	session.Stderr = errbuf

	writer := &Writer{
		procerr: make(chan error),
		writer:  w,
	}

	cmdline := "cat > " + shell.ReadableEscapeArg(path)

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
			// Bundle up stderr and hope it's useful
			err = fmt.Errorf("Command %#v on %#v: %w: %s",
				cmdline, host.connect, err, e)
		}

		writer.procerr <- err
		close(writer.procerr)
		session.Put()
	}()

	return writer, nil
}

func (host *Host) Remove(fpath string) error {
	session, err := host.pool.Get(host.connect)
	if err != nil {
		return err
	}

	outbuf := &bytes.Buffer{}
	errbuf := &bytes.Buffer{}

	session.Stdout = outbuf
	session.Stderr = errbuf

	cmdline := "rm -f " + shell.ReadableEscapeArg(fpath)

	if err := session.Run(cmdline); err != nil {
		e := strings.TrimSpace(errbuf.String())

		if strings.HasPrefix(e, "rm: ") && strings.HasSuffix(e, "No such file or directory") {
			// emulate os.Stat
			return &os.PathError{
				Op:   "rm",
				Path: fpath,
				Err:  syscall.ENOENT,
			}
		} else {
			// Bundle up stderr and hope it's useful
			err = fmt.Errorf("Command %#v on host %#v failed with %w: %s",
				cmdline, host.connect, err, e)
		}

		return err
	}

	return nil
}
