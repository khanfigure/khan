package rio

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"syscall"

	"github.com/keegancsmith/shell"
	"golang.org/x/crypto/ssh"
)

type SSHReader struct {
	config  *Config
	session *ssh.Session
	reader  *io.PipeReader
}

func (r *SSHReader) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}
func (r *SSHReader) Close() error {
	return r.reader.Close()
}

func (config *Config) ReadFile(path string) ([]byte, error) {
	if config.Pool == nil {
		return ioutil.ReadFile(path)
	}

	buf := &bytes.Buffer{}
	fh, err := config.Open(path)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(buf, fh); err != nil {
		return nil, err
	}
	if err := fh.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (config *Config) Open(path string) (io.ReadCloser, error) {
	if config.Pool == nil {
		return os.Open(path)
	}

	reader := &SSHReader{
		config: config,
	}

	session, err := config.Pool.Get(config.Host)
	if err != nil {
		return nil, err
	}

	r, w := io.Pipe()

	errbuf := &bytes.Buffer{}

	session.Stdout = w
	session.Stderr = errbuf
	reader.reader = r

	cmd := "cat " + shell.ReadableEscapeArg(path)
	if config.Sudo != "" {
		cmd = "sudo -u " + shell.ReadableEscapeArg(config.Sudo) + " " + cmd
	}

	if err := session.Start(cmd); err != nil {
		w.CloseWithError(err)
		session.Put()
		return nil, err
	}

	go func() {
		err := session.Wait()
		e := strings.TrimSpace(errbuf.String())

		if strings.HasPrefix(e, "cat: ") && strings.HasSuffix(e, "No such file or directory") {
			// emulate os.Open
			err = &os.PathError{
				Op:   "open",
				Path: path,
				Err:  syscall.ENOENT,
			}
		} else {
			errbuf.WriteTo(os.Stderr)
		}
		w.CloseWithError(err)
		session.Put()
	}()

	return reader, nil
}

type SSHWriter struct {
	config  *Config
	session *ssh.Session
	writer  *io.PipeWriter
}

func (w *SSHWriter) Write(p []byte) (int, error) {
	return w.writer.Write(p)
}
func (w *SSHWriter) Close() error {
	return w.writer.Close()
}

func (config *Config) Create(path string) (io.WriteCloser, error) {
	if config.Pool == nil {
		return os.Create(path)
	}

	writer := &SSHWriter{
		config: config,
	}

	session, err := config.Pool.Get(config.Host)
	if err != nil {
		return nil, err
	}

	r, w := io.Pipe()

	session.Stdin = r
	session.Stderr = os.Stderr
	writer.writer = w

	cmd := "cat > " + shell.ReadableEscapeArg(path)
	if config.Sudo != "" {
		cmd = "sudo -u " + shell.ReadableEscapeArg(config.Sudo) + " " + cmd
	}

	if err := session.Start(cmd); err != nil {
		r.CloseWithError(err)
		session.Put()
		return nil, err
	}

	go func() {
		r.CloseWithError(session.Wait())
		session.Put()
	}()

	return writer, nil
}

func (config *Config) Remove(path string) error {
	if config.Pool == nil {
		return os.Remove(path)
	}

	session, err := config.Pool.Get(config.Host)
	if err != nil {
		return err
	}

	outbuf := &bytes.Buffer{}
	errbuf := &bytes.Buffer{}

	session.Stdout = outbuf
	session.Stderr = errbuf

	cmdline := "rm " + shell.ReadableEscapeArg(path)
	if config.Sudo != "" {
		cmdline = "sudo -u " + shell.ReadableEscapeArg(config.Sudo) + " " + cmdline
	}

	if config.Verbose {
		fmt.Println("ssh", config.Host, cmdline)
	}

	if err := session.Run(cmdline); err != nil {
		if config.Verbose {
			fmt.Println("ssh", config.Host, cmdline, err)
		}

		e := strings.TrimSpace(errbuf.String())

		if strings.HasPrefix(e, "rm: ") && strings.HasSuffix(e, "No such file or directory") {
			// emulate os.Stat
			return &os.PathError{
				Op:   "rm",
				Path: path,
				Err:  syscall.ENOENT,
			}
		}

		return err
	}

	return nil
}
