package rio

import (
	"fmt"
	"io"
	"os"
)

type Host interface {
	String() string
	Info() (*Info, error)

	TmpFile() (string, error)
	TmpDir() (string, error)
	Cleanup() error

	Exec(cmd *Cmd) error

	Stat(string) (os.FileInfo, error)
	Open(string) (io.ReadCloser, error)
	ReadFile(string) ([]byte, error)
	Create(string) (io.WriteCloser, error)
	Remove(string) error // I'd rather call this Delete. But in this case, follow "os" package style.
	Chmod(string, os.FileMode) error
	Chown(string, uint32, uint32) error
	Rename(string, string) error

	User(string) (*User, error)
	CreateUser(*User) error
	UpdateUser(*User) error
	DeleteUser(string) error

	Group(string) (*Group, error)
	CreateGroup(*Group) error
	UpdateGroup(*Group) error
	DeleteGroup(string) error
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
