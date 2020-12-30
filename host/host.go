package host

import (
	"io"
	"os"
)

type Host interface {
	String() string

	Create(string) (io.WriteCloser, error)
	Remove(string) error

	Open(string) (io.ReadCloser, error)
	ReadFile(string) ([]byte, error)
	Stat(string) (os.FileInfo, error)

	Info() (*Info, error)
}

type Info struct {
	// Host metadata usually extracted from uname command
	Uname    string
	Hostname string
	Kernel   string
	OS       string
	Arch     string
}
