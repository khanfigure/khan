package khan

/*
import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"syscall"

	"github.com/keegancsmith/shell"
	"golang.org/x/crypto/ssh"
)

type sshReader struct {
	host    *Host
	session *ssh.Session
	reader  *io.PipeReader

	closemu  sync.Mutex
	closed   bool
	closeerr error
	procerr  chan error
}

func (r *sshReader) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}
func (r *sshReader) Close() error {
	//fmt.Println("sshReader close()")

	r.closemu.Lock()
	if !r.closed {
		r.closed = true
		cerr := r.reader.Close()
		r.closeerr = <-r.procerr
		if r.closeerr == nil && cerr != nil {
			r.closeerr = cerr
		}
	}
	r.closemu.Unlock()

	//fmt.Println("sshReader close():", r.closeerr)
	return r.closeerr
}

func (host *Host) ReadFile(path string) ([]byte, error) {
	if !host.SSH {
		return ioutil.ReadFile(path)
	}

	buf := &bytes.Buffer{}

	//fmt.Println("ReadFile open")
	fh, err := host.Open(path)
	//fmt.Println("ReadFile open:", err)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	//fmt.Println("ReadFile copy")
	_, err = io.Copy(buf, fh)
	//fmt.Println("ReadFile copy:", n, err)
	if err != nil {
		return nil, err
	}
	//fmt.Println("ReadFile close")
	err = fh.Close()
	//fmt.Println("ReadFile close:", err)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (host *Host) Open(path string) (io.ReadCloser, error) {
	if !host.SSH {
		return os.Open(path)
	}

	reader := &sshReader{
		host:    host,
		procerr: make(chan error),
	}

	session, err := host.Run.Pool.Get(host.Host)
	if err != nil {
		return nil, err
	}

	r, w := io.Pipe()

	errbuf := &bytes.Buffer{}

	session.Stdout = w
	session.Stderr = errbuf
	reader.reader = r

	cmdline := "cat " + shell.ReadableEscapeArg(path)

	if host.Run.Verbose {
		//fmt.Println("sshexec", host.Host, cmdline)
	}

	if err := session.Start(cmdline); err != nil {
		w.Close()
		r.Close()
		session.Put()
		return nil, err
	}

	go func() {
		//fmt.Println("sshexec", host.Host, cmdline, "waiting for process finish")
		err := session.Wait()
		if host.Run.Verbose {
			//fmt.Println("sshexec", host.Host, cmdline, err)
		}
		//fmt.Println("sshexec", host.Host, cmdline, "waiting for process finish done:", err)
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
					cmdline, host.Host, err, e)
			}
		}

		// This will let blocked reads finish
		w.CloseWithError(err)

		reader.procerr <- err
		//fmt.Println("sshexec", host.Host, cmdline, "error to reader:", err)
		close(reader.procerr)
		session.Put()
		//fmt.Println("sshexec", host.Host, cmdline, "put()")
	}()

	return reader, nil
}

type SSHWriter struct {
	host    *Host
	session *ssh.Session
	writer  *io.PipeWriter

	closemu  sync.Mutex
	closed   bool
	closeerr error
	procerr  chan error
}

func (w *SSHWriter) Write(p []byte) (int, error) {
	return w.writer.Write(p)
}
func (w *SSHWriter) Close() error {
	//fmt.Println("SSHWriter close()")

	w.closemu.Lock()
	if !w.closed {
		w.closed = true
		cerr := w.writer.Close()
		w.closeerr = <-w.procerr
		if w.closeerr == nil && cerr != nil {
			w.closeerr = cerr
		}
	}
	w.closemu.Unlock()

	//fmt.Println("SSHWriter close():", w.closeerr)
	return w.closeerr
}

func (host *Host) Create(path string) (io.WriteCloser, error) {
	if !host.SSH {
		return os.Create(path)
	}

	writer := &SSHWriter{
		host:    host,
		procerr: make(chan error),
	}

	session, err := host.Run.Pool.Get(host.Host)
	if err != nil {
		return nil, err
	}

	r, w := io.Pipe()

	errbuf := &bytes.Buffer{}

	session.Stdin = r
	session.Stderr = errbuf
	writer.writer = w

	cmdline := "cat > " + shell.ReadableEscapeArg(path)

	if host.Run.Verbose {
		//fmt.Println("sshexec", host.Host, cmdline)
	}

	if err := session.Start(cmdline); err != nil {
		w.Close()
		r.Close()
		session.Put()
		return nil, err
	}

	go func() {
		//fmt.Println("sshexec", host.Host, cmdline, "waiting for process finish")
		err := session.Wait()
		if host.Run.Verbose {
			//fmt.Println("sshexec", host.Host, cmdline, err)
		}
		e := strings.TrimSpace(errbuf.String())

		if err != nil {
			// Bundle up stderr and hope it's useful
			err = fmt.Errorf("Command %#v on host %#v: %w: %s",
				cmdline, host.Host, err, e)
		}

		//fmt.Println("sshexec", host.Host, cmdline, "waiting for process finish done:", err)
		writer.procerr <- err
		//fmt.Println("sshexec", host.Host, cmdline, "error sent to writer")
		close(writer.procerr)

		//r.CloseWithError(err)

		session.Put()
	}()

	return writer, nil
}

func (host *Host) Remove(path string) error {
	if !host.SSH {
		return os.Remove(path)
	}

	session, err := host.Run.Pool.Get(host.Host)
	if err != nil {
		return err
	}

	outbuf := &bytes.Buffer{}
	errbuf := &bytes.Buffer{}

	session.Stdout = outbuf
	session.Stderr = errbuf

	cmdline := "rm -f " + shell.ReadableEscapeArg(path)

	if host.Run.Verbose {
		//fmt.Println("sshexec", host.Host, cmdline)
	}

	if err := session.Run(cmdline); err != nil {
		if host.Run.Verbose {
			//fmt.Println("sshexec", host.Host, cmdline, err)
		}

		e := strings.TrimSpace(errbuf.String())

		if strings.HasPrefix(e, "rm: ") && strings.HasSuffix(e, "No such file or directory") {
			// emulate os.Stat
			return &os.PathError{
				Op:   "rm",
				Path: path,
				Err:  syscall.ENOENT,
			}
		} else {
			// Bundle up stderr and hope it's useful
			err = fmt.Errorf("Command %#v on host %#v failed with %w: %s",
				cmdline, host.Host, err, e)
		}

		return err
	}

	return nil
}*/
