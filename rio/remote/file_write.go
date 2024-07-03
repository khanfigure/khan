package remote

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"khan.rip/rio/util"

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
	if host.verbose {
		log.Println(host, ">", path)
	}

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
	return util.Remove(host, fpath)
}

func (host *Host) Rename(oldpath, newpath string) error {
	return util.Rename(host, oldpath, newpath)
}

func (host *Host) Chown(fpath string, uid uint32, gid uint32) error {
	return util.Chown(host, fpath, uid, gid)
}

func (host *Host) Chmod(fpath string, perms os.FileMode) error {
	return util.Chmod(host, fpath, perms)
}
