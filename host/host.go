package host

import (
	"fmt"
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
	Uname    string
	Hostname string
	Kernel   string
	OS       string
	Arch     string
}

func (info *Info) String() string {
	return fmt.Sprintf("%s (%s/%s)", info.Hostname, info.OS, info.Arch)
}
